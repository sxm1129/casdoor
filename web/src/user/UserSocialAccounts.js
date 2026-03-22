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
import {Card, Table, Tag, Empty} from "antd";
import {LinkOutlined, DisconnectOutlined} from "@ant-design/icons";
import i18next from "i18next";

class UserSocialAccounts extends React.Component {
  render() {
    const {user, userIdentities} = this.props;
    if (!user) {
      return null;
    }

    const identities = userIdentities || [];

    const columns = [
      {
        title: i18next.t("general:Provider"),
        dataIndex: "providerDisplayName",
        key: "providerDisplayName",
        render: (text, record) => (
          <span>
            <LinkOutlined style={{marginRight: 8}} />
            {text || record.providerName}
          </span>
        ),
      },
      {
        title: i18next.t("user:Provider type"),
        dataIndex: "providerType",
        key: "providerType",
        render: (text) => <Tag>{text}</Tag>,
      },
      {
        title: i18next.t("user:Identity ID"),
        dataIndex: "idpId",
        key: "idpId",
        ellipsis: true,
      },
      {
        title: i18next.t("user:Username"),
        dataIndex: "username",
        key: "username",
      },
      {
        title: i18next.t("general:Status"),
        key: "status",
        render: () => <Tag color="success">{i18next.t("user:Connected")}</Tag>,
      },
    ];

    return (
      <Card title={i18next.t("user:Social Accounts")} bordered={false}>
        {identities.length > 0 ? (
          <Table
            dataSource={identities}
            columns={columns}
            rowKey={(record) => `${record.providerName}_${record.idpId}`}
            pagination={false}
            size="small"
          />
        ) : (
          <Empty
            image={<DisconnectOutlined style={{fontSize: 48, color: "#d9d9d9"}} />}
            description={i18next.t("user:No social accounts connected")}
          />
        )}
      </Card>
    );
  }
}

export default UserSocialAccounts;
