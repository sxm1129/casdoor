// Copyright 2026 The Casdoor Authors. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

import React from "react";
import {Card, Tag, Table, Empty} from "antd";
import i18next from "i18next";

class UserRolePermissions extends React.Component {
  render() {
    const {user} = this.props;
    if (!user) {
      return null;
    }

    const roleColumns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        render: (text) => <Tag color="blue">{text}</Tag>,
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
      },
      {
        title: i18next.t("general:Owner"),
        dataIndex: "owner",
        key: "owner",
      },
    ];

    const permissionColumns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
        render: (text) => <Tag color="green">{text}</Tag>,
      },
      {
        title: i18next.t("permission:Resource type"),
        dataIndex: "resourceType",
        key: "resourceType",
      },
      {
        title: i18next.t("permission:Effect"),
        dataIndex: "effect",
        key: "effect",
        render: (text) => (
          <Tag color={text === "Allow" ? "success" : "error"}>{text}</Tag>
        ),
      },
    ];

    return (
      <div>
        <Card
          title={i18next.t("user:Roles")}
          bordered={false}
          style={{marginBottom: 16}}
        >
          {user.roles && user.roles.length > 0 ? (
            <Table
              dataSource={user.roles}
              columns={roleColumns}
              rowKey="name"
              pagination={false}
              size="small"
            />
          ) : (
            <Empty description={i18next.t("general:No data")} />
          )}
        </Card>

        <Card title={i18next.t("user:Permissions")} bordered={false}>
          {user.permissions && user.permissions.length > 0 ? (
            <Table
              dataSource={user.permissions}
              columns={permissionColumns}
              rowKey="name"
              pagination={false}
              size="small"
            />
          ) : (
            <Empty description={i18next.t("general:No data")} />
          )}
        </Card>
      </div>
    );
  }
}

export default UserRolePermissions;
