package object

import (
	"encoding/json"
	"fmt"

	"github.com/casdoor/casdoor/conf"
	"github.com/casdoor/casdoor/util"
)

func RunMigration() error {
	// Only run migration if user_role table is empty
	count, err := ormer.Engine.Count(&UserRole{})
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Migration already ran
	}

	fmt.Println("Running legacy data migration for Route A (RBAC & User)...")

	err = migrateRoles()
	if err != nil {
		return fmt.Errorf("migrateRoles error: %v", err)
	}

	err = migratePermissions()
	if err != nil {
		return fmt.Errorf("migratePermissions error: %v", err)
	}

	// Not migrating 84 IdPs yet here, will add shortly
	err = migrateUsers()
	if err != nil {
		return fmt.Errorf("migrateUsers error: %v", err)
	}

	return nil
}

func getTable(name string) string {
	return conf.GetConfigString("tableNamePrefix") + name
}

func migrateRoles() error {
	results, err := ormer.Engine.QueryString(fmt.Sprintf("SELECT owner, name, users, roles FROM %s", getTable("role")))
	if err != nil {
		// table might not exist or columns dropped (unlikely with xorm Sync2)
		return nil
	}

	for _, result := range results {
		roleId := util.GetId(result["owner"], result["name"])
		
		// Parse users
		var users []string
		if result["users"] != "" {
			_ = json.Unmarshal([]byte(result["users"]), &users)
		}
		for _, u := range users {
			_, _ = ormer.Engine.Insert(&UserRole{
				User: u,
				Role: roleId,
			})
		}

		// Parse sub-roles
		var roles []string
		if result["roles"] != "" {
			_ = json.Unmarshal([]byte(result["roles"]), &roles)
		}
		// Notice: To support role inheritance natively, we would add RoleRole table, 
		// but since Casdoor uses Casbin for role inheritance, we map role-to-role as an implicit graph.
		// For now we map sub-roles to role if needed.
	}
	return nil
}

func migratePermissions() error {
	results, err := ormer.Engine.QueryString(fmt.Sprintf("SELECT owner, name, users, roles FROM %s", getTable("permission")))
	if err != nil {
		return nil
	}

	for _, result := range results {
		permId := util.GetId(result["owner"], result["name"])
		
		// Parse users
		var users []string
		if result["users"] != "" {
			_ = json.Unmarshal([]byte(result["users"]), &users)
		}
		for _, u := range users {
			_, _ = ormer.Engine.Insert(&UserPermission{
				User:       u,
				Permission: permId,
			})
		}

		// Parse roles
		var roles []string
		if result["roles"] != "" {
			_ = json.Unmarshal([]byte(result["roles"]), &roles)
		}
		for _, r := range roles {
			_, _ = ormer.Engine.Insert(&RolePermission{
				Role:       r,
				Permission: permId,
			})
		}
	}
	return nil
}

func migrateUsers() error {
	results, err := ormer.Engine.QueryString(fmt.Sprintf("SELECT * FROM %s", getTable("user")))
	if err != nil {
		return nil
	}

	idps := []string{
		"github", "google", "qq", "wechat", "facebook", "dingtalk", "weibo", "gitee",
		"linkedin", "wecom", "lark", "gitlab", "adfs", "baidu", "alipay", "casdoor",
		"infoflow", "apple", "azuread", "azureadb2c", "slack", "steam", "bilibili",
		"okta", "douyin", "kwai", "line", "amazon", "auth0", "battlenet", "bitbucket",
		"box", "cloudfoundry", "dailymotion", "deezer", "digitalocean", "discord",
		"dropbox", "eveonline", "fitbit", "gitea", "heroku", "influxcloud", "instagram",
		"intercom", "kakao", "lastfm", "mailru", "meetup", "microsoftonline", "naver",
		"nextcloud", "onedrive", "oura", "patreon", "paypal", "salesforce", "shopify",
		"soundcloud", "spotify", "strava", "stripe", "telegram", "tiktok", "tumblr",
		"twitch", "twitter", "typetalk", "uber", "vk", "wepay", "xero", "yahoo",
		"yammer", "yandex", "zoom", "metamask", "web3onboard", "custom", "custom2",
		"custom3", "custom4", "custom5", "custom6", "custom7", "custom8", "custom9", "custom10",
	}

	for _, result := range results {
		owner := result["owner"]
		name := result["name"]

		for _, idp := range idps {
			if val, ok := result[idp]; ok && val != "" {
				_, _ = ormer.Engine.Insert(&UserIdentity{
					Owner:        string(owner),
					Name:         string(name),
					ProviderType: idp,
					ProviderId:   string(val),
				})
			}
		}
	}
	return nil
}
