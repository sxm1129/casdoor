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
import {Avatar, Badge, Button, Card, Form, Input, List, Modal, Space, Spin, Table, Tabs, Tag, Typography, message} from "antd";
import {DeleteOutlined, ExclamationCircleOutlined, HistoryOutlined, KeyOutlined, LockOutlined, PlusOutlined, UserOutlined} from "@ant-design/icons";

const {Text, Paragraph} = Typography;
import * as Setting from "../Setting";
import i18next from "i18next";
import "./portal.css";

class PortalLayout extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      user: null,
      sessions: [],
      apiKeys: [],
      newKeyVisible: false,
      newKeyValue: null,
      loading: true,
      activeTab: "profile",
      passwordForm: {oldPassword: "", newPassword: "", confirmPassword: ""},
      savingPassword: false,
    };
  }

  componentDidMount() {
    this.fetchProfile();
    this.fetchSessions();
    this.fetchApiKeys();
  }

  fetchProfile() {
    fetch(`${Setting.ServerUrl}/api/get-portal-profile`, {
      method: "GET",
      credentials: "include",
      headers: {"Accept-Language": Setting.getAcceptLanguage()},
    })
      .then(res => res.json())
      .then(res => {
        if (res.status === "ok") {
          this.setState({user: res.data, loading: false});
        } else {
          this.setState({loading: false});
          Setting.showMessage("error", res.msg);
        }
      });
  }

  fetchSessions() {
    fetch(`${Setting.ServerUrl}/api/get-portal-sessions`, {
      method: "GET",
      credentials: "include",
      headers: {"Accept-Language": Setting.getAcceptLanguage()},
    })
      .then(res => res.json())
      .then(res => {
        if (res.status === "ok") {
          this.setState({sessions: res.data || []});
        }
      });
  }

  fetchApiKeys() {
    fetch(`${Setting.ServerUrl}/api/get-api-keys`, {
      method: "GET",
      credentials: "include",
      headers: {"Accept-Language": Setting.getAcceptLanguage()},
    })
      .then(res => res.json())
      .then(res => {
        if (res.status === "ok") {
          this.setState({apiKeys: res.data || []});
        }
      });
  }

  generateApiKey() {
    fetch(`${Setting.ServerUrl}/api/add-api-key`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": Setting.getAcceptLanguage(),
      },
      body: JSON.stringify({description: "Portal generated key"}),
    })
      .then(res => res.json())
      .then(res => {
        if (res.status === "ok") {
          this.setState({newKeyVisible: true, newKeyValue: res.data});
          this.fetchApiKeys();
        } else {
          message.error(res.msg);
        }
      });
  }

  deleteApiKey(id) {
    Modal.confirm({
      title: i18next.t("general:Sure to delete") + "?",
      icon: <ExclamationCircleOutlined />,
      onOk: () => {
        fetch(`${Setting.ServerUrl}/api/delete-api-key`, {
          method: "POST",
          credentials: "include",
          headers: {
            "Content-Type": "application/json",
            "Accept-Language": Setting.getAcceptLanguage(),
          },
          body: JSON.stringify({id}),
        })
          .then(res => res.json())
          .then(res => {
            if (res.status === "ok") {
              message.success(i18next.t("general:Successfully deleted"));
              this.fetchApiKeys();
            } else {
              message.error(res.msg);
            }
          });
      },
    });
  }

  changePassword() {
    const {passwordForm} = this.state;
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      message.error(i18next.t("user:Two passwords you typed do not match"));
      return;
    }
    if (!passwordForm.newPassword) {
      message.error(i18next.t("user:Please input your password"));
      return;
    }
    this.setState({savingPassword: true});
    fetch(`${Setting.ServerUrl}/api/set-password`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": Setting.getAcceptLanguage(),
      },
      body: JSON.stringify({
        userOwner: this.state.user?.owner,
        userName: this.state.user?.name,
        oldPassword: passwordForm.oldPassword,
        newPassword: passwordForm.newPassword,
      }),
    })
      .then(res => res.json())
      .then(res => {
        this.setState({savingPassword: false});
        if (res.status === "ok") {
          message.success(i18next.t("user:Password changed successfully"));
          this.setState({passwordForm: {oldPassword: "", newPassword: "", confirmPassword: ""}});
        } else {
          message.error(res.msg);
        }
      });
  }

  updateProfile(field, value) {
    this.setState(prevState => ({
      user: {...prevState.user, [field]: value},
    }));
  }

  saveProfile() {
    fetch(`${Setting.ServerUrl}/api/update-portal-profile`, {
      method: "POST",
      credentials: "include",
      headers: {
        "Content-Type": "application/json",
        "Accept-Language": Setting.getAcceptLanguage(),
      },
      body: JSON.stringify(this.state.user),
    })
      .then(res => res.json())
      .then(res => {
        if (res.status === "ok") {
          message.success(i18next.t("general:Successfully saved"));
        } else {
          message.error(res.msg);
        }
      });
  }

  renderProfile() {
    const {user} = this.state;
    if (!user) {return null;}

    return (
      <Card bordered={false}>
        <div style={{textAlign: "center", marginBottom: 24}}>
          <Avatar size={80} icon={<UserOutlined />} src={user.avatar} />
          <h3 style={{marginTop: 12}}>{user.displayName || user.name}</h3>
          <p style={{color: "#888"}}>{user.email}</p>
        </div>
        <Form layout="vertical">
          <Form.Item label={i18next.t("general:Display name")}>
            <Input value={user.displayName} onChange={e => this.updateProfile("displayName", e.target.value)} />
          </Form.Item>
          <Form.Item label={i18next.t("general:Email")}>
            <Input value={user.email} onChange={e => this.updateProfile("email", e.target.value)} />
          </Form.Item>
          <Form.Item label={i18next.t("general:Phone")}>
            <Input value={user.phone} onChange={e => this.updateProfile("phone", e.target.value)} />
          </Form.Item>
          <Form.Item label={i18next.t("user:Location")}>
            <Input value={user.location} onChange={e => this.updateProfile("location", e.target.value)} />
          </Form.Item>
          <Form.Item label={i18next.t("user:Bio")}>
            <Input.TextArea value={user.bio} rows={3} onChange={e => this.updateProfile("bio", e.target.value)} />
          </Form.Item>
          <Button type="primary" onClick={() => this.saveProfile()}>
            {i18next.t("general:Save")}
          </Button>
        </Form>
      </Card>
    );
  }

  renderSessions() {
    const columns = [
      {title: i18next.t("general:Name"), dataIndex: "name", key: "name"},
      {title: i18next.t("general:Application"), dataIndex: "application", key: "application"},
      {title: i18next.t("general:Created time"), dataIndex: "createdTime", key: "createdTime",
        render: (text) => Setting.getFormattedDate(text)},
    ];

    return (
      <Card title={i18next.t("portal:Active Sessions")} bordered={false}>
        <Table dataSource={this.state.sessions} columns={columns} rowKey="name" size="small" pagination={false} />
      </Card>
    );
  }

  renderSecurity() {
    const {passwordForm, savingPassword, user} = this.state;
    return (
      <div>
        <Card title={i18next.t("user:Change Password")} bordered={false} style={{marginBottom: 24}}>
          <Form layout="vertical" style={{maxWidth: 400}}>
            <Form.Item label={i18next.t("user:Old Password")}>
              <Input.Password
                value={passwordForm.oldPassword}
                onChange={e => this.setState({passwordForm: {...passwordForm, oldPassword: e.target.value}})}
              />
            </Form.Item>
            <Form.Item label={i18next.t("user:New Password")}>
              <Input.Password
                value={passwordForm.newPassword}
                onChange={e => this.setState({passwordForm: {...passwordForm, newPassword: e.target.value}})}
              />
            </Form.Item>
            <Form.Item label={i18next.t("user:Confirm Password")}>
              <Input.Password
                value={passwordForm.confirmPassword}
                onChange={e => this.setState({passwordForm: {...passwordForm, confirmPassword: e.target.value}})}
              />
            </Form.Item>
            <Button type="primary" loading={savingPassword} onClick={() => this.changePassword()}>
              {i18next.t("user:Change Password")}
            </Button>
          </Form>
        </Card>
        <Card title={i18next.t("mfa:Multi-factor authentication")} bordered={false}>
          <List
            dataSource={[
              {type: "Email", enabled: user?.mfaPhoneEnabled || false},
              {type: "SMS", enabled: user?.mfaEmailEnabled || false},
              {type: "TOTP", enabled: !!user?.totpSecret},
            ]}
            renderItem={item => (
              <List.Item>
                <Space>
                  <Text strong>{item.type}</Text>
                  {item.enabled
                    ? <Badge status="success" text={i18next.t("general:Enabled")} />
                    : <Badge status="default" text={i18next.t("general:Disabled")} />
                  }
                </Space>
              </List.Item>
            )}
          />
        </Card>
      </div>
    );
  }

  renderApiKeys() {
    const columns = [
      {title: "ID", dataIndex: "id", key: "id", width: 80},
      {title: i18next.t("general:Description"), dataIndex: "description", key: "description"},
      {title: i18next.t("general:Created time"), dataIndex: "createdTime", key: "createdTime",
        render: (text) => Setting.getFormattedDate(text)},
      {title: "Prefix", dataIndex: "prefix", key: "prefix",
        render: (text) => <Tag>{text}...</Tag>},
      {title: i18next.t("general:Action"), key: "action",
        render: (_, record) => (
          <Button danger size="small" icon={<DeleteOutlined />} onClick={() => this.deleteApiKey(record.id)}>
            {i18next.t("general:Delete")}
          </Button>
        )},
    ];

    return (
      <div>
        <Card
          title="API Keys"
          bordered={false}
          extra={
            <Button type="primary" icon={<PlusOutlined />} onClick={() => this.generateApiKey()}>
              {i18next.t("general:Generate")}
            </Button>
          }
        >
          <Table dataSource={this.state.apiKeys} columns={columns} rowKey="id" size="small" pagination={false} />
        </Card>
        <Modal
          title="New API Key Generated"
          open={this.state.newKeyVisible}
          onOk={() => this.setState({newKeyVisible: false, newKeyValue: null})}
          onCancel={() => this.setState({newKeyVisible: false, newKeyValue: null})}
          cancelButtonProps={{style: {display: "none"}}}
        >
          <p style={{color: "#ff4d4f", fontWeight: 600, marginBottom: 12}}>
            Copy this key now. It will not be shown again.
          </p>
          <Paragraph copyable style={{background: "#f5f5f5", padding: 12, borderRadius: 8, fontFamily: "monospace"}}>
            {this.state.newKeyValue}
          </Paragraph>
        </Modal>
      </div>
    );
  }

  render() {
    if (this.state.loading) {
      return <div style={{textAlign: "center", padding: 100}}><Spin size="large" /></div>;
    }

    const tabItems = [
      {key: "profile", label: <span><UserOutlined />{i18next.t("portal:Profile")}</span>, children: this.renderProfile()},
      {key: "security", label: <span><LockOutlined />{i18next.t("portal:Security")}</span>, children: this.renderSecurity()},
      {key: "sessions", label: <span><HistoryOutlined />{i18next.t("portal:Sessions")}</span>, children: this.renderSessions()},
      {key: "api-keys", label: <span><KeyOutlined />API Keys</span>, children: this.renderApiKeys()},
    ];

    return (
      <div className="portal-container">
        <div className="portal-card">
          <h2 className="portal-title">{i18next.t("portal:My Account")}</h2>
          <Tabs
            activeKey={this.state.activeTab}
            onChange={key => this.setState({activeTab: key})}
            items={tabItems}
          />
        </div>
      </div>
    );
  }
}

export default PortalLayout;
