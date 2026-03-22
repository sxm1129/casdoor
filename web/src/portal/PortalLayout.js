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
import {Card, Tabs, Form, Input, Button, Avatar, Descriptions, Table, Tag, Spin, message} from "antd";
import {UserOutlined, LockOutlined, LinkOutlined, HistoryOutlined} from "@ant-design/icons";
import * as Setting from "../Setting";
import i18next from "i18next";
import "./portal.css";

class PortalLayout extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      user: null,
      sessions: [],
      loading: true,
      activeTab: "profile",
    };
  }

  componentDidMount() {
    this.fetchProfile();
    this.fetchSessions();
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

  render() {
    if (this.state.loading) {
      return <div style={{textAlign: "center", padding: 100}}><Spin size="large" /></div>;
    }

    const tabItems = [
      {key: "profile", label: <span><UserOutlined />{i18next.t("portal:Profile")}</span>, children: this.renderProfile()},
      {key: "sessions", label: <span><HistoryOutlined />{i18next.t("portal:Sessions")}</span>, children: this.renderSessions()},
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
