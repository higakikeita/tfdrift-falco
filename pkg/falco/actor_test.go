package falco

import (
	"testing"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestResolveActor(t *testing.T) {
	tests := []struct {
		name     string
		in       types.UserIdentity
		wantUser string
	}{
		{
			name:     "IAM user keeps its userName",
			in:       types.UserIdentity{Type: "IAMUser", UserName: "alice", ARN: "arn:aws:iam::123:user/alice"},
			wantUser: "alice",
		},
		{
			name: "AssumedRole/SSO resolves to the session name (the human)",
			in: types.UserIdentity{
				Type: "AssumedRole",
				ARN:  "arn:aws:sts::123456789012:assumed-role/AWSReservedSSO_Admin_abc123/keita@example.com",
			},
			wantUser: "keita@example.com",
		},
		{
			name: "AssumedRole (plain role assumption) resolves to the session name",
			in: types.UserIdentity{
				Type: "AssumedRole",
				ARN:  "arn:aws:sts::123456789012:assumed-role/DeployRole/ci-pipeline-42",
			},
			wantUser: "ci-pipeline-42",
		},
		{
			name: "AssumedRole with no session falls back to the role name",
			in: types.UserIdentity{
				Type: "AssumedRole",
				ARN:  "arn:aws:sts::123456789012:assumed-role/DeployRole",
			},
			wantUser: "DeployRole",
		},
		{
			name: "empty userName + principalId session is used",
			in: types.UserIdentity{
				Type:        "AssumedRole",
				PrincipalID: "AROAEXAMPLEID:session-bob",
			},
			wantUser: "session-bob",
		},
		{
			name: "root falls back to the ARN tail rather than staying blank",
			in: types.UserIdentity{
				Type: "Root",
				ARN:  "arn:aws:iam::123456789012:root",
			},
			wantUser: "root",
		},
		{
			name:     "no identity at all stays empty (nothing to invent)",
			in:       types.UserIdentity{Type: "AWSService"},
			wantUser: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveActor(tt.in)
			assert.Equal(t, tt.wantUser, got.UserName,
				"resolveActor must surface a human 'who' for %s", tt.name)
		})
	}
}

func TestParseAssumedRoleARN(t *testing.T) {
	role, session, ok := parseAssumedRoleARN("arn:aws:sts::123:assumed-role/MyRole/my-session")
	assert.True(t, ok)
	assert.Equal(t, "MyRole", role)
	assert.Equal(t, "my-session", session)

	_, _, ok = parseAssumedRoleARN("arn:aws:iam::123:user/alice")
	assert.False(t, ok, "a plain IAM user ARN is not an assumed-role ARN")
}

func TestIsSSORole(t *testing.T) {
	assert.True(t, IsSSORole("AWSReservedSSO_AdministratorAccess_1a2b3c"))
	assert.False(t, IsSSORole("DeployRole"))
}
