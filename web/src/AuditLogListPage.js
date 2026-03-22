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
import {Table, Tag, Card, Input, Select, DatePicker, Space, Tooltip, Modal} from "antd";
import {EyeOutlined, SearchOutlined} from "@ant-design/icons";
import * as Setting from "./Setting";
import i18next from "i18next";

const {Option} = Select;

class AuditLogListPage extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      auditLogs: [],
      total: 0,
      loading: false,
      page: 1,
      pageSize: 20,
      targetTypeFilter: "",
      searchValue: "",
      selectedLog: null,
      detailVisible: false,
    };
  }

  componentDidMount() {
    this.fetchAuditLogs();
  }

  fetchAuditLogs() {
    this.setState({loading: true});
    const {page, pageSize, targetTypeFilter, searchValue} = this.state;
    const field = targetTypeFilter ? "target_type" : "";
    const value = targetTypeFilter || searchValue;

    fetch(`${Setting.ServerUrl}/api/get-audit-logs?pageSize=${pageSize}&p=${page}&field=${field}&value=${value}`, {
      method: "GET",
      credentials: "include",
      headers: {"Accept-Language": Setting.getAcceptLanguage()},
    })
      .then(res => res.json())
      .then(res => {
        if (res.status === "ok") {
          this.setState({
            auditLogs: res.data || [],
            total: res.data2 || 0,
            loading: false,
          });
        } else {
          Setting.showMessage("error", res.msg);
          this.setState({loading: false});
        }
      });
  }

  renderFieldChanges(fieldChangesStr) {
    if (!fieldChangesStr || fieldChangesStr === "[]") {
      return <Tag>No changes</Tag>;
    }

    try {
      const changes = JSON.parse(fieldChangesStr);
      return (
        <div>
          {changes.map((change, index) => (
            <div key={index} style={{marginBottom: 8, padding: "8px 12px", background: "#fafafa", borderRadius: 6}}>
              <strong style={{color: "#1890ff"}}>{change.field}</strong>
              <div style={{fontSize: 12, color: "#999", marginTop: 4}}>
                <span style={{color: "#ff4d4f"}}>- {String(change.oldValue).substring(0, 100)}</span>
                <br />
                <span style={{color: "#52c41a"}}>+ {String(change.newValue).substring(0, 100)}</span>
              </div>
            </div>
          ))}
        </div>
      );
    } catch (e) {
      return <span>{fieldChangesStr.substring(0, 200)}</span>;
    }
  }

  render() {
    const columns = [
      {
        title: i18next.t("general:Time"),
        dataIndex: "createdTime",
        key: "createdTime",
        width: 180,
        sorter: true,
        render: (text) => Setting.getFormattedDate(text),
      },
      {
        title: i18next.t("audit:Actor"),
        dataIndex: "actor",
        key: "actor",
        width: 150,
        render: (text) => <Tag color="blue">{text}</Tag>,
      },
      {
        title: i18next.t("audit:Action"),
        dataIndex: "action",
        key: "action",
        width: 100,
        render: (text) => {
          const colorMap = {create: "success", update: "processing", delete: "error"};
          return <Tag color={colorMap[text] || "default"}>{text}</Tag>;
        },
      },
      {
        title: i18next.t("audit:Target type"),
        dataIndex: "targetType",
        key: "targetType",
        width: 120,
      },
      {
        title: i18next.t("audit:Target ID"),
        dataIndex: "targetId",
        key: "targetId",
        width: 200,
        ellipsis: true,
      },
      {
        title: i18next.t("audit:IP"),
        dataIndex: "actorIp",
        key: "actorIp",
        width: 130,
      },
      {
        title: i18next.t("general:Action"),
        key: "detail",
        width: 80,
        render: (_, record) => (
          <Tooltip title={i18next.t("general:View")}>
            <EyeOutlined
              style={{cursor: "pointer", color: "#1890ff"}}
              onClick={() => this.setState({selectedLog: record, detailVisible: true})}
            />
          </Tooltip>
        ),
      },
    ];

    return (
      <div>
        <Card
          title={i18next.t("audit:Audit Logs")}
          extra={
            <Space>
              <Select
                placeholder={i18next.t("audit:Target type")}
                allowClear
                style={{width: 150}}
                value={this.state.targetTypeFilter || undefined}
                onChange={v => this.setState({targetTypeFilter: v || ""}, () => this.fetchAuditLogs())}
              >
                <Option value="user">User</Option>
                <Option value="role">Role</Option>
                <Option value="permission">Permission</Option>
                <Option value="organization">Organization</Option>
              </Select>
              <Input.Search
                placeholder={i18next.t("general:Search")}
                style={{width: 200}}
                onSearch={v => this.setState({searchValue: v}, () => this.fetchAuditLogs())}
              />
            </Space>
          }
        >
          <Table
            dataSource={this.state.auditLogs}
            columns={columns}
            rowKey="id"
            loading={this.state.loading}
            pagination={{
              current: this.state.page,
              pageSize: this.state.pageSize,
              total: this.state.total,
              showSizeChanger: true,
              showTotal: (total) => `${i18next.t("general:Total")} ${total}`,
              onChange: (page, pageSize) => this.setState({page, pageSize}, () => this.fetchAuditLogs()),
            }}
            size="middle"
          />
        </Card>

        <Modal
          title={i18next.t("audit:Audit Log Detail")}
          open={this.state.detailVisible}
          onCancel={() => this.setState({detailVisible: false})}
          footer={null}
          width={600}
        >
          {this.state.selectedLog && (
            <div>
              <p><strong>{i18next.t("audit:Actor")}:</strong> {this.state.selectedLog.actor}</p>
              <p><strong>{i18next.t("audit:Action")}:</strong> <Tag>{this.state.selectedLog.action}</Tag></p>
              <p><strong>{i18next.t("audit:Target")}:</strong> {this.state.selectedLog.targetType} / {this.state.selectedLog.targetId}</p>
              <p><strong>{i18next.t("audit:IP")}:</strong> {this.state.selectedLog.actorIp}</p>
              <p><strong>{i18next.t("audit:Login method")}:</strong> {this.state.selectedLog.loginMethod || "-"}</p>
              <p><strong>{i18next.t("audit:Request URI")}:</strong> {this.state.selectedLog.requestUri || "-"}</p>
              <hr />
              <p><strong>{i18next.t("audit:Field changes")}:</strong></p>
              {this.renderFieldChanges(this.state.selectedLog.fieldChanges)}
            </div>
          )}
        </Modal>
      </div>
    );
  }
}

export default AuditLogListPage;
