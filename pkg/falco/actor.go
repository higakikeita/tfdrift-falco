package falco

import (
	"strings"

	"github.com/keitahigaki/tfdrift-falco/pkg/types"
)

// resolveActor enriches a CloudTrail userIdentity so the human "who" is
// populated even for AssumedRole / SSO callers.
//
// For IAMUser / Root, CloudTrail sets userName (ct.user) directly. For
// AssumedRole — which is how SSO (AWS IAM Identity Center) and any
// role-assumption call appears, i.e. the common enterprise case — userName is
// empty; the human identifier lives in the ARN's session name
// (arn:aws:sts::<acct>:assumed-role/<Role>/<session>) or in the principalId
// (<AROA...>:<session>). Without this, TFDrift-Falco's headline "who changed
// it" is blank exactly when it matters most (pus #7 in ADR-0012).
func resolveActor(ui types.UserIdentity) types.UserIdentity {
	if ui.UserName != "" {
		return ui
	}

	if role, session, ok := parseAssumedRoleARN(ui.ARN); ok {
		switch {
		case session != "":
			ui.UserName = session
		default:
			ui.UserName = role
		}
		return ui
	}

	// principalId for an assumed role looks like "AROAEXAMPLEID:session-name".
	if i := strings.LastIndex(ui.PrincipalID, ":"); i >= 0 && i+1 < len(ui.PrincipalID) {
		ui.UserName = ui.PrincipalID[i+1:]
		return ui
	}

	// Last resort: the final ARN segment (e.g. the user/role name, or "root"),
	// so the field is never silently empty when any identity is available.
	// Splits on both "/" and ":" to handle "…:user/alice" and "…:root".
	if ui.ARN != "" {
		if i := strings.LastIndexAny(ui.ARN, "/:"); i >= 0 && i+1 < len(ui.ARN) {
			ui.UserName = ui.ARN[i+1:]
		} else {
			ui.UserName = ui.ARN
		}
	}

	return ui
}

// parseAssumedRoleARN extracts the role and session name from an STS
// assumed-role ARN: arn:aws:sts::<account>:assumed-role/<RoleName>/<sessionName>.
// For SSO the RoleName is typically AWSReservedSSO_<PermissionSet>_<hash> and
// the sessionName is the user's identity (often their email).
func parseAssumedRoleARN(arn string) (role, session string, ok bool) {
	const marker = ":assumed-role/"
	idx := strings.Index(arn, marker)
	if idx < 0 {
		return "", "", false
	}
	rest := arn[idx+len(marker):]
	parts := strings.SplitN(rest, "/", 2)
	role = parts[0]
	if len(parts) == 2 {
		session = parts[1]
	}
	return role, session, true
}

// IsSSORole reports whether a role name is an AWS IAM Identity Center (SSO)
// reserved role, so callers can label the actor's origin.
func IsSSORole(roleName string) bool {
	return strings.HasPrefix(roleName, "AWSReservedSSO_")
}
