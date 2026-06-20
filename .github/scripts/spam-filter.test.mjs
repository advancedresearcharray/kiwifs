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
