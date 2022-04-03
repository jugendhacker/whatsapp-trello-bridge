package server

import "time"

type WebhookRequest struct {
	Action struct {
		ID              string `json:"id"`
		IDMemberCreator string `json:"idMemberCreator"`
		Data            struct {
			Board struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"board"`
			Card struct {
				IDShort int    `json:"idShort"`
				Name    string `json:"name"`
				ID      string `json:"id"`
			} `json:"card"`
			Voted bool `json:"voted"`
		} `json:"data"`
		Type          string    `json:"type"`
		Date          time.Time `json:"date"`
		MemberCreator struct {
			ID         string `json:"id"`
			AvatarHash string `json:"avatarHash"`
			FullName   string `json:"fullName"`
			Initials   string `json:"initials"`
			Username   string `json:"username"`
		} `json:"memberCreator"`
	} `json:"action"`
	Model struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Desc           string `json:"desc"`
		Closed         bool   `json:"closed"`
		IDOrganization string `json:"idOrganization"`
		Pinned         bool   `json:"pinned"`
		URL            string `json:"url"`
		Prefs          struct {
			PermissionLevel string `json:"permissionLevel"`
			Voting          string `json:"voting"`
			Comments        string `json:"comments"`
			Invitations     string `json:"invitations"`
			SelfJoin        bool   `json:"selfJoin"`
			CardCovers      bool   `json:"cardCovers"`
			CanBePublic     bool   `json:"canBePublic"`
			CanBeOrg        bool   `json:"canBeOrg"`
			CanBePrivate    bool   `json:"canBePrivate"`
			CanInvite       bool   `json:"canInvite"`
		} `json:"prefs"`
		LabelNames struct {
			Yellow string `json:"yellow"`
			Red    string `json:"red"`
			Purple string `json:"purple"`
			Orange string `json:"orange"`
			Green  string `json:"green"`
			Blue   string `json:"blue"`
		} `json:"labelNames"`
	} `json:"model"`
}
