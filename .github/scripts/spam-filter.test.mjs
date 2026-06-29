import assert from 'node:assert/strict';
import test from 'node:test';
import {
  SPAM_LOG_ISSUE,
  buildLogCommentBody,
  cjkRatio,
  ensureIssueUnlocked,
  isCjkDominant,
  logSpamModeration,
  runSpamFilter,
  sanitizeSnippet,
} from './spam-filter.cjs';

test('cjkRatio returns 0 for empty body', () => {
  assert.equal(cjkRatio(''), 0);
  assert.equal(cjkRatio('   '), 0);
});

test('cjkRatio detects CJK-dominant content', () => {
  const body = '这个仓库的情况只是冰山一角。如果我们继续对刷星行为视而不见';
  assert.ok(isCjkDominant(body));
  assert.ok(cjkRatio(body) >= 0.5);
});

test('cjkRatio allows English-dominant content', () => {
  const body = 'This is a normal English comment about KiwiFS features.';
  assert.equal(isCjkDominant(body), false);
});

test('sanitizeSnippet redacts tokens emails and urls', () => {
  const raw = 'Contact me at user@example.com with ghp_abc123secret https://evil.test/x';
  const sanitized = sanitizeSnippet(raw);
  assert.match(sanitized, /\[REDACTED_EMAIL\]/);
  assert.match(sanitized, /\[REDACTED_TOKEN\]/);
  assert.match(sanitized, /\[REDACTED_URL\]/);
  assert.doesNotMatch(sanitized, /user@example.com/);
});

test('buildLogCommentBody includes moderation metadata', () => {
  const body = buildLogCommentBody({
    maintainer: 'amelia751',
    action: 'issue',
    author: 'binybow623',
    issueNumber: 394,
    cjkRatioValue: 0.83,
    isComment: false,
    snippet: 'spam preview',
    bodyLength: 250,
  });

  assert.match(body, /@amelia751/);
  assert.match(body, /binybow623/);
  assert.match(body, /#394/);
  assert.match(body, /83%/);
  assert.match(body, /\.\.\./);
  assert.match(body, /sanitized/i);
});

test('ensureIssueUnlocked unlocks locked moderation log issue', async () => {
  const calls = [];
  const github = {
    rest: {
      issues: {
        get: async () => {
          calls.push('get');
          return { data: { locked: true } };
        },
        unlock: async () => {
          calls.push('unlock');
        },
      },
    },
  };

  const unlocked = await ensureIssueUnlocked(github, {
    owner: 'kiwifs',
    repo: 'kiwifs',
    issueNumber: SPAM_LOG_ISSUE,
  });

  assert.equal(unlocked, true);
  assert.deepEqual(calls, ['get', 'unlock']);
});

test('ensureIssueUnlocked is a no-op when issue is already unlocked', async () => {
  const calls = [];
  const github = {
    rest: {
      issues: {
        get: async () => {
          calls.push('get');
          return { data: { locked: false } };
        },
        unlock: async () => {
          calls.push('unlock');
        },
      },
    },
  };

  const unlocked = await ensureIssueUnlocked(github, {
    owner: 'kiwifs',
    repo: 'kiwifs',
    issueNumber: SPAM_LOG_ISSUE,
  });

  assert.equal(unlocked, false);
  assert.deepEqual(calls, ['get']);
});

test('logSpamModeration unlocks locked tracking issue before commenting', async () => {
  const calls = [];
  const github = {
    rest: {
      issues: {
        get: async () => {
          calls.push('get');
          return { data: { locked: true } };
        },
        unlock: async () => {
          calls.push('unlock');
        },
        createComment: async () => {
          calls.push('createComment');
        },
      },
    },
  };

  await logSpamModeration(
    github,
    { repo: { owner: 'kiwifs', repo: 'kiwifs' } },
    {
      author: 'binybow623',
      issueNumber: 394,
      cjkRatioValue: 0.83,
      isComment: false,
      body: '这个仓库的情况只是冰山一角。',
    },
  );

  assert.deepEqual(calls, ['get', 'unlock', 'createComment']);
});

test('logSpamModeration skips when issue is the moderation log target', async () => {
  let createCommentCalled = false;
  const github = {
    rest: {
      issues: {
        get: async () => ({ data: { locked: false } }),
        createComment: async () => {
          createCommentCalled = true;
        },
      },
    },
  };

  await logSpamModeration(
    github,
    { repo: { owner: 'kiwifs', repo: 'kiwifs' } },
    {
      author: 'binybow623',
      issueNumber: SPAM_LOG_ISSUE,
      cjkRatioValue: 0.83,
      isComment: false,
      body: 'spam body',
    },
  );

  assert.equal(createCommentCalled, false);
});

test('logSpamModeration does not throw when comment creation still fails', async () => {
  const github = {
    rest: {
      issues: {
        get: async () => ({ data: { locked: false } }),
        unlock: async () => {
          throw new Error('should not unlock');
        },
        createComment: async () => {
          throw new Error('Unable to create comment because issue is locked.');
        },
      },
    },
  };

  await assert.doesNotReject(async () => {
    await logSpamModeration(
      github,
      { repo: { owner: 'kiwifs', repo: 'kiwifs' } },
      {
        author: 'binybow623',
        issueNumber: 394,
        cjkRatioValue: 0.83,
        isComment: false,
        body: 'spam body',
      },
    );
  });
});

test('runSpamFilter skips moderation log issue #392', async () => {
  let createCommentCalled = false;
  const github = {
    rest: {
      issues: {
        createComment: async () => {
          createCommentCalled = true;
        },
      },
    },
  };

  await runSpamFilter({
    github,
    context: {
      repo: { owner: 'kiwifs', repo: 'kiwifs' },
      payload: {
        issue: {
          number: SPAM_LOG_ISSUE,
          body: '这个仓库的情况只是冰山一角。',
          user: { login: 'binybow623' },
        },
      },
    },
  });

  assert.equal(createCommentCalled, false);
});

test('runSpamFilter skips comments on moderation log issue #392', async () => {
  let graphqlCalled = false;
  const github = {
    graphql: async () => {
      graphqlCalled = true;
    },
    rest: {
      issues: {
        listComments: async () => ({ data: [] }),
      },
      repos: {
        listContributors: async () => ({ data: [] }),
      },
    },
  };

  await runSpamFilter({
    github,
    context: {
      repo: { owner: 'kiwifs', repo: 'kiwifs' },
      payload: {
        issue: {
          number: SPAM_LOG_ISSUE,
          user: { login: 'binybow623' },
        },
        comment: {
          id: 1,
          node_id: 'C_1',
          body: '这个仓库的情况只是冰山一角。',
          user: { login: 'binybow623' },
        },
      },
    },
  });

  assert.equal(graphqlCalled, false);
});

test('runSpamFilter uses issue-scoped comment listing for prior activity', async () => {
  let listedIssue = null;
  const github = {
    rest: {
      repos: {
        getCollaboratorPermissionLevel: async () => {
          throw new Error('not collaborator');
        },
        listContributors: async () => ({ data: [] }),
      },
      orgs: {
        checkMembershipForUser: async () => {
          throw new Error('not member');
        },
      },
      issues: {
        listComments: async ({ issue_number }) => {
          listedIssue = issue_number;
          return { data: [{ id: 99, user: { login: 'binybow623' } }] };
        },
        update: async () => {},
        lock: async () => {},
        get: async () => ({ data: { locked: false } }),
        createComment: async () => {},
      },
    },
    graphql: async () => {},
  };

  await runSpamFilter({
    github,
    context: {
      repo: { owner: 'kiwifs', repo: 'kiwifs' },
      payload: {
        issue: {
          number: 500,
          body: '这个仓库的情况只是冰山一角。',
          user: { login: 'binybow623' },
        },
      },
    },
  });

  assert.equal(listedIssue, 500);
});
