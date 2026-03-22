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
import {Button, Card, Input, Modal, Space, Table, Tag, Tooltip, message} from "antd";
import {CopyOutlined, DeleteOutlined, KeyOutlined, PlusOutlined} from "@ant-design/icons";
import * as Setting from "./Setting";
import i18next from "i18next";
import copy from "copy-to-clipboard";

class ApiKeyListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      apiKeys: [],
      loading: false,
      createModalVisible: false,
      newKeyName: "",
      newKeyDisplayName: "",
      createdRawKey: null,
    };
  }

  componentDidMount() {
    this.fetchApiKeys();
  }

  fetchApiKeys() {
    this.setState({loading: true});
    const owner = this.props.account?.owner || "admin";
    fetch(`${Setting.ServerUrl}/api/get-api-keys?owner=${owner}`, {
      method: "GET",
      credentials: "include",
      headers: {"Accept-Language": Setting.getAcceptLanguage()},
    })
      .then(res => res.json())
      .then(res => {
        if (res.status === "ok") {
          this.setState({apiKeys: res.data || [], loading: false});
        } else {
          this.setState({loading: false});
        }
      });
  }

  createApiKey() {
    const owner = this.props.account?.owner || "admin";
    const body = {
      owner: owner,
      name: this.state.newKeyName,
      displayName: this.state.newKeyDisplayName,
    };

    fetch(`${Setting.ServerUrl}/api/add-api-key`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": Setting.getAcceptLanguage(),
      },
      body: JSON.stringify(body),
    })
      .then(res => res.json())
      .then(res => {
        if (res.status === "ok") {
          this.setState({
            createdRawKey: res.data,
            createModalVisible: false,
            newKeyName: "",
            newKeyDisplayName: "",
          });
          this.fetchApiKeys();
          message.success(i18next.t("general:Successfully added"));
        } else {
          message.error(res.msg);
        }
      });
  }

  deleteApiKey(record) {
    fetch(`${Setting.ServerUrl}/api/delete-api-key`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": Setting.getAcceptLanguage(),
      },
      body: JSON.stringify(record),
    })
      .then(res => res.json())
      .then(res => {
        if (res.status === "ok") {
          this.fetchApiKeys();
          message.success(i18next.t("general:Successfully deleted"));
        } else {
          message.error(res.msg);
        }
      });
  }

  render() {
    const columns = [
      {
        title: i18next.t("general:Name"),
        dataIndex: "name",
        key: "name",
      },
      {
        title: i18next.t("general:Display name"),
        dataIndex: "displayName",
        key: "displayName",
      },
      {
        title: i18next.t("api_key:Key prefix"),
        dataIndex: "keyPrefix",
        key: "keyPrefix",
        render: (text) => <code>{text}...</code>,
      },
      {
        title: i18next.t("general:Status"),
        dataIndex: "isEnabled",
        key: "isEnabled",
        render: (enabled) => (
          <Tag color={enabled ? "success" : "default"}>
            {enabled ? i18next.t("general:Enabled") : i18next.t("general:Disabled")}
          </Tag>
        ),
      },
      {
        title: i18next.t("api_key:Last used"),
        dataIndex: "lastUsedAt",
        key: "lastUsedAt",
        render: (text) => text ? Setting.getFormattedDate(text) : "-",
      },
      {
        title: i18next.t("general:Created time"),
        dataIndex: "createdTime",
        key: "createdTime",
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("general:Action"),
        key: "action",
        render: (_, record) => (
          <Tooltip title={i18next.t("general:Delete")}>
            <Button
              danger
              type="text"
              icon={<DeleteOutlined />}
              onClick={() => {
                Modal.confirm({
                  title: i18next.t("general:Sure to delete") + "?",
                  onOk: () => this.deleteApiKey(record),
                });
              }}
            />
          </Tooltip>
        ),
      },
    ];

    return (
      <div>
        <Card
          title={
            <Space>
              <KeyOutlined />
              {i18next.t("api_key:API Keys")}
            </Space>
          }
          extra={
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={() => this.setState({createModalVisible: true})}
            >
              {i18next.t("general:Add")}
            </Button>
          }
        >
          <Table
            dataSource={this.state.apiKeys}
            columns={columns}
            rowKey="name"
            loading={this.state.loading}
            pagination={false}
            size="middle"
          />
        </Card>

        {/* Create API Key Modal */}
        <Modal
          title={i18next.t("api_key:Create API Key")}
          open={this.state.createModalVisible}
          onOk={() => this.createApiKey()}
          onCancel={() => this.setState({createModalVisible: false})}
          okText={i18next.t("general:OK")}
        >
          <div style={{marginBottom: 16}}>
            <label>{i18next.t("general:Name")}</label>
            <Input
              value={this.state.newKeyName}
              onChange={e => this.setState({newKeyName: e.target.value})}
              placeholder="my-api-key"
            />
          </div>
          <div>
            <label>{i18next.t("general:Display name")}</label>
            <Input
              value={this.state.newKeyDisplayName}
              onChange={e => this.setState({newKeyDisplayName: e.target.value})}
              placeholder="My API Key"
            />
          </div>
        </Modal>

        {/* Show Created Key (once) */}
        <Modal
          title={i18next.t("api_key:API Key Created")}
          open={this.state.createdRawKey !== null}
          onOk={() => this.setState({createdRawKey: null})}
          onCancel={() => this.setState({createdRawKey: null})}
          footer={[
            <Button key="copy" type="primary" icon={<CopyOutlined />}
              onClick={() => {copy(this.state.createdRawKey); message.success("Copied!");}}>
              {i18next.t("general:Copy")}
            </Button>,
            <Button key="close" onClick={() => this.setState({createdRawKey: null})}>
              {i18next.t("general:Close")}
            </Button>,
          ]}
        >
          <p style={{color: "#ff4d4f", fontWeight: "bold", marginBottom: 8}}>
            {i18next.t("api_key:This key will only be shown once. Please copy and store it securely.")}
          </p>
          <Input.TextArea
            value={this.state.createdRawKey}
            rows={3}
            readOnly
            style={{fontFamily: "monospace", fontSize: 12}}
          />
        </Modal>
      </div>
    );
  }
}

export default ApiKeyListPage;
