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

function sanitizeSnippet(body, maxLen = 200) {
  return body
    .substring(0, maxLen)
    .replace(/\n/g, ' ')
    .replace(/(?:ghp_|gho_|github_pat_)[A-Za-z0-9_]+/g, '[REDACTED_TOKEN]')
    .replace(/[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}/g, '[REDACTED_EMAIL]')
    .replace(/https?:\/\/[^\s]+/g, '[REDACTED_URL]');
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
    '**Content preview (sanitized):**',
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

  if (issueNumber === SPAM_LOG_ISSUE) {
    console.log(`Skipping moderation log for #${SPAM_LOG_ISSUE} (log target)`);
    return;
  }

  const snippet = sanitizeSnippet(body);
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
    console.error(`Failed to log spam to #${SPAM_LOG_ISSUE}: ${error.message}`);
  }
}

function isTrustedAuthor(author) {
  return TRUSTED_BOTS.includes(author) || author === MAINTAINER;
}

async function hasElevatedAccess(github, context, author) {
  try {
    const { data: permLevel } = await github.rest.repos.getCollaboratorPermissionLevel({
      owner: context.repo.owner,
      repo: context.repo.repo,
      username: author,
    });
    if (['admin', 'write', 'maintain'].includes(permLevel.permission)) {
      return true;
    }
  } catch (error) {
    console.log(`Permission check for ${author}: ${error.message}`);
  }

  try {
    await github.rest.orgs.checkMembershipForUser({
      org: context.repo.owner,
      username: author,
    });
    return true;
  } catch (error) {
    console.log(`Org membership check for ${author}: ${error.message}`);
  }

  return false;
}

async function hasPriorActivity(github, context, author, issueNumber, commentId) {
  try {
    const { data: comments } = await github.rest.issues.listComments({
      owner: context.repo.owner,
      repo: context.repo.repo,
      issue_number: issueNumber,
      per_page: 20,
    });
    return comments.some(
      (comment) => comment.user.login === author && comment.id !== commentId,
    );
  } catch (error) {
    console.log(`Prior activity check on #${issueNumber}: ${error.message}`);
    return false;
  }
}

async function isRepoContributor(github, context, author) {
  try {
    const { data: contributors } = await github.rest.repos.listContributors({
      owner: context.repo.owner,
      repo: context.repo.repo,
      per_page: 100,
    });
    return contributors.some((contributor) => contributor.login === author);
  } catch (error) {
    console.log(`Contributor check for ${author}: ${error.message}`);
    return false;
  }
}

async function runSpamFilter({ github, context }) {
  try {
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

    if (isTrustedAuthor(author)) {
      return;
    }

    if (await hasElevatedAccess(github, context, author)) {
      return;
    }

    const ratio = cjkRatio(body);
    if (ratio < 0.5) {
      return;
    }

    const commentId = context.payload.comment?.id;
    if (
      (await hasPriorActivity(github, context, author, issueNumber, commentId)) ||
      (await isRepoContributor(github, context, author))
    ) {
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
      console.error(`Failed to block ${author}: ${error.message}`);
    }

    await logSpamModeration(github, context, {
      author,
      issueNumber,
      cjkRatioValue: ratio,
      isComment,
      body,
    });
  } catch (error) {
    console.error(`runSpamFilter failed: ${error.message}`);
    throw error;
  }
}

module.exports = {
  SPAM_LOG_ISSUE,
  MAINTAINER,
  TRUSTED_BOTS,
  cjkRatio,
  isCjkDominant,
  sanitizeSnippet,
  buildLogCommentBody,
  ensureIssueUnlocked,
  logSpamModeration,
  runSpamFilter,
};
