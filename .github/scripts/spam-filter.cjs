'use strict';

const SPAM_LOG_ISSUE = 392;
const MAINTAINER = 'amelia751';

const TRUSTED_BOTS = [
  'github-actions[bot]',
  'dependabot[bot]',
  'release-please[bot]',
  'cursor[bot]',
  'renovate[bot]',
];

const CJK_REGEX =
  /[\u4e00-\u9fff\u3400-\u4dbf\u3000-\u303f\uff00-\uffef\u2e80-\u2eff\u3200-\u32ff\ufe30-\ufe4f]/g;

function cjkRatio(body) {
  const cjkMatches = body.match(CJK_REGEX) || [];
  const totalChars = body.replace(/\s/g, '').length;
  if (totalChars === 0) {
    return 0;
  }
  return cjkMatches.length / totalChars;
}

function isCjkDominant(body, threshold = 0.5) {
  return cjkRatio(body) >= threshold;
}

function buildLogCommentBody({
  maintainer,
  action,
  author,
  issueNumber,
  cjkRatioValue,
  isComment,
  snippet,
  bodyLength,
}) {
  return [
    `@${maintainer} 🚨 **Spam ${action} hidden**`,
    '',
    '| Field | Value |',
    '|-------|-------|',
    `| Author | \`${author}\` |`,
    `| Issue/PR | #${issueNumber} |`,
    `| CJK ratio | ${(cjkRatioValue * 100).toFixed(0)}% |`,
    `| Action taken | ${isComment ? 'Comment minimized' : 'Issue closed + locked'} + user blocked |`,
    '',
    '**Content preview:**',
    `> ${snippet}${bodyLength > 200 ? '...' : ''}`,
  ].join('\n');
}

async function ensureIssueUnlocked(github, { owner, repo, issueNumber }) {
  const { data: issue } = await github.rest.issues.get({
    owner,
    repo,
    issue_number: issueNumber,
  });

  if (!issue.locked) {
    return false;
  }

  await github.rest.issues.unlock({
    owner,
    repo,
    issue_number: issueNumber,
  });

  return true;
}

async function logSpamModeration(github, context, details) {
  const owner = context.repo.owner;
  const repo = context.repo.repo;
  const {
    author,
    issueNumber,
    cjkRatioValue,
    isComment,
    body,
  } = details;

  const snippet = body.substring(0, 200).replace(/\n/g, ' ');
  const action = isComment ? 'comment' : 'issue';
  const commentBody = buildLogCommentBody({
    maintainer: MAINTAINER,
    action,
    author,
    issueNumber,
    cjkRatioValue,
    isComment,
    snippet,
    bodyLength: body.length,
  });

  try {
    const unlocked = await ensureIssueUnlocked(github, {
      owner,
      repo,
      issueNumber: SPAM_LOG_ISSUE,
    });
    if (unlocked) {
      console.log(`Unlocked #${SPAM_LOG_ISSUE} for spam moderation logging`);
    }

    await github.rest.issues.createComment({
      owner,
      repo,
      issue_number: SPAM_LOG_ISSUE,
      body: commentBody,
    });
  } catch (error) {
    console.log(`Failed to log spam to #${SPAM_LOG_ISSUE}: ${error.message}`);
  }
}

async function runSpamFilter({ github, context }) {
  const isComment = !!context.payload.comment;
  const body = isComment
    ? context.payload.comment.body
    : context.payload.issue.body || '';
  const author = isComment
    ? context.payload.comment.user.login
    : context.payload.issue.user.login;
  const issueNumber = context.payload.issue.number;

  if (issueNumber === SPAM_LOG_ISSUE) {
    console.log(`Skipping spam filter on moderation log issue #${SPAM_LOG_ISSUE}`);
    return;
  }

  if (TRUSTED_BOTS.includes(author) || author === MAINTAINER) {
    return;
  }

  try {
    const { data: permLevel } = await github.rest.repos.getCollaboratorPermissionLevel({
      owner: context.repo.owner,
      repo: context.repo.repo,
      username: author,
    });
    if (['admin', 'write', 'maintain'].includes(permLevel.permission)) {
      return;
    }
  } catch (error) {
    // Not a collaborator — continue with validation.
  }

  try {
    await github.rest.orgs.checkMembershipForUser({
      org: context.repo.owner,
      username: author,
    });
    return;
  } catch (error) {
    // Not an org member — continue.
  }

  const ratio = cjkRatio(body);
  if (ratio < 0.5) {
    return;
  }

  let hasPriorActivity = false;
  try {
    const { data: comments } = await github.rest.issues.listCommentsForRepo({
      owner: context.repo.owner,
      repo: context.repo.repo,
      per_page: 5,
      sort: 'created',
      direction: 'desc',
    });
    hasPriorActivity = comments.some(
      (comment) =>
        comment.user.login === author &&
        comment.id !== (context.payload.comment?.id),
    );
  } catch (error) {
    // Best-effort prior activity check.
  }

  let isContributor = false;
  try {
    const { data: contributors } = await github.rest.repos.listContributors({
      owner: context.repo.owner,
      repo: context.repo.repo,
      per_page: 100,
    });
    isContributor = contributors.some((contributor) => contributor.login === author);
  } catch (error) {
    // Best-effort contributor check.
  }

  if (hasPriorActivity || isContributor) {
    return;
  }

  console.log(
    `🚨 Spam detected from ${author} on #${issueNumber} (CJK ratio: ${(ratio * 100).toFixed(0)}%)`,
  );

  if (isComment) {
    const commentNodeId = context.payload.comment.node_id;
    await github.graphql(
      `
        mutation($id: ID!) {
          minimizeComment(input: { subjectId: $id, classifier: SPAM }) {
            minimizedComment { isMinimized }
          }
        }
      `,
      { id: commentNodeId },
    );
  } else {
    await github.rest.issues.update({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      state: 'closed',
      state_reason: 'not_planned',
    });
    await github.rest.issues.lock({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      lock_reason: 'spam',
    });
  }

  try {
    await github.rest.orgs.blockUser({
      org: context.repo.owner,
      username: author,
    });
  } catch (error) {
    console.log(`Failed to block ${author}: ${error.message}`);
  }

  await logSpamModeration(github, context, {
    author,
    issueNumber,
    cjkRatioValue: ratio,
    isComment,
    body,
  });
}

module.exports = {
  SPAM_LOG_ISSUE,
  MAINTAINER,
  TRUSTED_BOTS,
  cjkRatio,
  isCjkDominant,
  buildLogCommentBody,
  ensureIssueUnlocked,
  logSpamModeration,
  runSpamFilter,
};
