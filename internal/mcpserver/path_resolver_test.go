package mcpserver

import "testing"

func TestResolveMCPPathReadOnlyAcceptsDecodedCJK(t *testing.T) {
	got, err := resolveMCPPath("00 Inbox/GitHub Issue 기반 에이전트 지시 게이트웨이 설계 초안.md", mcpPathReadOnly)
	if err != nil {
		t.Fatalf("resolveMCPPath returned error: %v", err)
	}
	want := "00 Inbox/GitHub Issue 기반 에이전트 지시 게이트웨이 설계 초안.md"
	if got.Path != want {
		t.Fatalf("path = %q, want %q", got.Path, want)
	}
}

func TestResolveMCPPathReadOnlyDecodesPageURL(t *testing.T) {
	input := "https://wiki.illuwa.click/page/00%20Inbox/GitHub%20Issue%20%EA%B8%B0%EB%B0%98%20%EC%97%90%EC%9D%B4%EC%A0%84%ED%8A%B8%20%EC%A7%80%EC%8B%9C%20%EA%B2%8C%EC%9D%B4%ED%8A%B8%EC%9B%A8%EC%9D%B4%20%EC%84%A4%EA%B3%84%20%EC%B4%88%EC%95%88.md"
	got, err := resolveMCPPath(input, mcpPathReadOnly)
	if err != nil {
		t.Fatalf("resolveMCPPath returned error: %v", err)
	}
	want := "00 Inbox/GitHub Issue 기반 에이전트 지시 게이트웨이 설계 초안.md"
	if got.Path != want {
		t.Fatalf("path = %q, want %q", got.Path, want)
	}
}

func TestResolveMCPPathReadOnlyDecodesEncodedPath(t *testing.T) {
	input := "00%20Inbox/GitHub%20Issue%20%EA%B8%B0%EB%B0%98%20%EC%97%90%EC%9D%B4%EC%A0%84%ED%8A%B8%20%EC%A7%80%EC%8B%9C%20%EA%B2%8C%EC%9D%B4%ED%8A%B8%EC%9B%A8%EC%9D%B4%20%EC%84%A4%EA%B3%84%20%EC%B4%88%EC%95%88.md"
	got, err := resolveMCPPath(input, mcpPathReadOnly)
	if err != nil {
		t.Fatalf("resolveMCPPath returned error: %v", err)
	}
	want := "00 Inbox/GitHub Issue 기반 에이전트 지시 게이트웨이 설계 초안.md"
	if got.Path != want {
		t.Fatalf("path = %q, want %q", got.Path, want)
	}
}

func TestResolveMCPPathMutationRejectsURLAndEncodedPath(t *testing.T) {
	cases := []string{
		"https://wiki.illuwa.click/page/00%20Inbox/%ED%85%8C%EC%8A%A4%ED%8A%B8.md",
		"00%20Inbox/%ED%85%8C%EC%8A%A4%ED%8A%B8.md",
	}
	for _, tc := range cases {
		if _, err := resolveMCPPath(tc, mcpPathMutation); err == nil {
			t.Fatalf("resolveMCPPath(%q, mutation) returned nil error", tc)
		}
	}
}

func TestResolveMCPPathRejectsUnsafeSegments(t *testing.T) {
	cases := []string{"../secret.md", "00 Inbox/../secret.md", "/absolute.md", "00 Inbox//bad.md", `00 Inbox\\bad.md`}
	for _, tc := range cases {
		if _, err := resolveMCPPath(tc, mcpPathReadOnly); err == nil {
			t.Fatalf("resolveMCPPath(%q) returned nil error", tc)
		}
	}
}
