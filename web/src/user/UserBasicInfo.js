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
import {Card, Col, Form, Input, Row, Select, Upload, Avatar} from "antd";
import {UserOutlined, UploadOutlined} from "@ant-design/icons";
import * as Setting from "../Setting";
import i18next from "i18next";

const {Option} = Select;

class UserBasicInfo extends React.Component {
  render() {
    const {user, onUpdateUserField, application} = this.props;
    if (!user) {
      return null;
    }

    return (
      <Card title={i18next.t("user:Basic Information")} bordered={false}>
        <Row gutter={[24, 16]}>
          <Col span={4} style={{textAlign: "center"}}>
            <Avatar
              size={100}
              icon={<UserOutlined />}
              src={user.avatar}
              style={{marginBottom: 16}}
            />
          </Col>
          <Col span={20}>
            <Row gutter={[16, 12]}>
              <Col span={12}>
                <Form.Item label={i18next.t("general:Name")}>
                  <Input value={user.name} disabled />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={i18next.t("general:Display name")}>
                  <Input
                    value={user.displayName}
                    onChange={e => onUpdateUserField("displayName", e.target.value)}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={i18next.t("general:Email")}>
                  <Input
                    value={user.email}
                    onChange={e => onUpdateUserField("email", e.target.value)}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={i18next.t("general:Phone")}>
                  <Input
                    value={user.phone}
                    onChange={e => onUpdateUserField("phone", e.target.value)}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={i18next.t("user:Affiliation")}>
                  <Input
                    value={user.affiliation}
                    onChange={e => onUpdateUserField("affiliation", e.target.value)}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={i18next.t("user:Title")}>
                  <Input
                    value={user.title}
                    onChange={e => onUpdateUserField("title", e.target.value)}
                  />
                </Form.Item>
              </Col>
              <Col span={24}>
                <Form.Item label={i18next.t("user:Bio")}>
                  <Input.TextArea
                    value={user.bio}
                    rows={3}
                    onChange={e => onUpdateUserField("bio", e.target.value)}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={i18next.t("user:Location")}>
                  <Input
                    value={user.location}
                    onChange={e => onUpdateUserField("location", e.target.value)}
                  />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item label={i18next.t("user:Homepage")}>
                  <Input
                    value={user.homepage}
                    onChange={e => onUpdateUserField("homepage", e.target.value)}
                  />
                </Form.Item>
              </Col>
            </Row>
          </Col>
        </Row>
      </Card>
    );
  }
}

export default UserBasicInfo;
