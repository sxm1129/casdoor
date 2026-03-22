// Copyright 2023 The Casdoor Authors. All Rights Reserved.
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

import {Spin, Tour} from "antd";
import * as echarts from "echarts";
import i18next from "i18next";
import React from "react";
import * as DashboardBackend from "../backend/DashboardBackend";
import * as Setting from "../Setting";
import * as TourConfig from "../TourConfig";

const Dashboard = (props) => {
  const [dashboardData, setDashboardData] = React.useState(null);
  const [isTourVisible, setIsTourVisible] = React.useState(TourConfig.getTourVisible());
  const nextPathName = TourConfig.getNextUrl("home");
  const chartRef = React.useRef(null);

  React.useEffect(() => {
    window.addEventListener("storageTourChanged", handleTourChange);
    return () => window.removeEventListener("storageTourChanged", handleTourChange);
  }, []);

  React.useEffect(() => {
    window.addEventListener("storageOrganizationChanged", handleOrganizationChange);
    return () => window.removeEventListener("storageOrganizationChanged", handleOrganizationChange);
  }, [props.owner]);

  React.useEffect(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      props.history.push("/apps");
    }
  }, [props.account]);

  const getOrganizationName = () => {
    let organization = localStorage.getItem("organization") === "All" ? "" : localStorage.getItem("organization");
    if (!Setting.isAdminUser(props.account) && Setting.isLocalAdminUser(props.account)) {
      organization = props.account.owner;
    }
    return organization;
  };

  React.useEffect(() => {
    if (!Setting.isLocalAdminUser(props.account)) {
      return;
    }

    const organization = getOrganizationName();
    DashboardBackend.getDashboard(organization).then((res) => {
      if (res.status === "ok") {
        setDashboardData(res.data);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  }, [props.owner]);

  // AUDIT R4-B1/R4-B3 fix: move echarts init into useEffect to avoid null DOM and React anti-pattern
  React.useEffect(() => {
    if (dashboardData === null || !chartRef.current) {
      return;
    }

    // AUDIT R4-S5 fix: reuse existing instance to prevent memory leaks
    let myChart = echarts.getInstanceByDom(chartRef.current);
    if (!myChart) {
      myChart = echarts.init(chartRef.current);
    }

    const currentDate = new Date();
    const dateArray = [];
    for (let i = 30; i >= 0; i--) {
      const date = new Date(currentDate);
      date.setDate(date.getDate() - i);
      const month = parseInt(date.getMonth()) + 1;
      const day = parseInt(date.getDate());
      const formattedDate = `${month}-${day}`;
      dateArray.push(formattedDate);
    }

    const option = {
      title: {text: i18next.t("home:Past 30 Days")},
      tooltip: {trigger: "axis"},
      legend: {data: [
        i18next.t("general:Users"),
        i18next.t("application:Providers"),
        i18next.t("general:Applications"),
        i18next.t("general:Organizations"),
        i18next.t("general:Subscriptions"),
        i18next.t("general:Roles"),
        i18next.t("general:Groups"),
        i18next.t("general:Resources"),
        i18next.t("general:Certs"),
        i18next.t("general:Permissions"),
        i18next.t("general:Transactions"),
        i18next.t("general:Models"),
        i18next.t("general:Adapters"),
        i18next.t("general:Enforcers"),
      ], top: "10%"},
      grid: {left: "3%", right: "4%", bottom: "0", top: "30%", containLabel: true},
      xAxis: {type: "category", boundaryGap: false, data: dateArray},
      yAxis: {type: "value"},
      series: [
        {name: i18next.t("general:Organizations"), type: "line", data: dashboardData.organizationCounts},
        {name: i18next.t("general:Users"), type: "line", data: dashboardData.userCounts},
        {name: i18next.t("application:Providers"), type: "line", data: dashboardData.providerCounts},
        {name: i18next.t("general:Applications"), type: "line", data: dashboardData.applicationCounts},
        {name: i18next.t("general:Subscriptions"), type: "line", data: dashboardData.subscriptionCounts},
        {name: i18next.t("general:Roles"), type: "line", data: dashboardData.roleCounts},
        {name: i18next.t("general:Groups"), type: "line", data: dashboardData.groupCounts},
        {name: i18next.t("general:Resources"), type: "line", data: dashboardData.resourceCounts},
        {name: i18next.t("general:Certs"), type: "line", data: dashboardData.certCounts},
        {name: i18next.t("general:Permissions"), type: "line", data: dashboardData.permissionCounts},
        {name: i18next.t("general:Transactions"), type: "line", data: dashboardData.transactionCounts},
        {name: i18next.t("general:Models"), type: "line", data: dashboardData.modelCounts},
        {name: i18next.t("general:Adapters"), type: "line", data: dashboardData.adapterCounts},
        {name: i18next.t("general:Enforcers"), type: "line", data: dashboardData.enforcerCounts},
      ],
    };
    myChart.setOption(option);

    // Handle window resize for responsive charts
    const handleResize = () => myChart.resize();
    window.addEventListener("resize", handleResize);

    return () => {
      window.removeEventListener("resize", handleResize);
      myChart.dispose();
    };
  }, [dashboardData]);

  const handleTourChange = () => {
    setIsTourVisible(TourConfig.getTourVisible());
  };

  const handleOrganizationChange = () => {
    if (!Setting.isLocalAdminUser(props.account)) {
      return;
    }

    setDashboardData(null);

    const organization = getOrganizationName();
    DashboardBackend.getDashboard(organization).then((res) => {
      if (res.status === "ok") {
        setDashboardData(res.data);
      } else {
        Setting.showMessage("error", res.msg);
      }
    });
  };

  const setIsTourToLocal = () => {
    TourConfig.setIsTourVisible(false);
    setIsTourVisible(false);
  };

  const handleTourComplete = () => {
    if (nextPathName !== "") {
      props.history.push("/" + nextPathName);
      TourConfig.setIsTourVisible(true);
    }
  };

  const getSteps = () => {
    const steps = TourConfig.TourObj["home"];
    steps.map((item, index) => {
      item.target = () => document.getElementById(item.id) || null;
      if (index === steps.length - 1) {
        item.nextButtonProps = {
          children: TourConfig.getNextButtonChild(nextPathName),
        };
      }
    });
    return steps;
  };

  // AUDIT R4-B4 fix: compute real metrics from dashboardData instead of hardcoded mock values
  const getNewUsersToday = () => {
    if (!dashboardData?.userCounts) {return 0;}
    return dashboardData.userCounts[30] - dashboardData.userCounts[29];
  };

  const getNewUsers7d = () => {
    if (!dashboardData?.userCounts) {return 0;}
    return dashboardData.userCounts[30] - dashboardData.userCounts[23];
  };

  const renderMetricCards = () => {
    if (dashboardData === null) {
      return null;
    }

    return (
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-10 px-4">
        <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm transition-shadow hover:shadow-md">
          <div className="flex justify-between items-start mb-4">
            <span className="material-symbols-outlined text-blue-600 bg-blue-50 p-2 rounded-lg">group</span>
            <span className="text-[10px] font-bold text-emerald-600 bg-emerald-50 px-2 py-1 rounded-full">+{getNewUsers7d()} (7d)</span>
          </div>
          <h3 className="text-xs font-bold text-gray-500 uppercase tracking-widest">{i18next.t("home:Total users")}</h3>
          <p className="text-3xl font-bold mt-1 text-gray-900">{dashboardData.userCounts[30].toLocaleString()}</p>
          <div className="w-full bg-gray-100 mt-4 h-1 rounded-full overflow-hidden">
            <div className="bg-blue-600 h-full" style={{width: "75%"}}></div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm transition-shadow hover:shadow-md">
          <div className="flex justify-between items-start mb-4">
            <span className="material-symbols-outlined text-blue-900 bg-blue-50 p-2 rounded-lg">apps</span>
            <span className="text-[10px] font-bold text-gray-500 bg-gray-100 px-2 py-1 rounded-full">{i18next.t("general:Stable")}</span>
          </div>
          <h3 className="text-xs font-bold text-gray-500 uppercase tracking-widest">{i18next.t("general:Applications")}</h3>
          <p className="text-3xl font-bold mt-1 text-gray-900">{dashboardData.applicationCounts[30]}</p>
          <div className="flex gap-1 mt-4">
            <div className="h-1 flex-1 bg-blue-600 rounded-full"></div>
            <div className="h-1 flex-1 bg-blue-600 rounded-full"></div>
            <div className="h-1 flex-1 bg-blue-600 rounded-full"></div>
            <div className="h-1 flex-1 bg-gray-100 rounded-full"></div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm transition-shadow hover:shadow-md">
          <div className="flex justify-between items-start mb-4">
            <span className="material-symbols-outlined text-red-600 bg-red-50 p-2 rounded-lg">trending_up</span>
          </div>
          <h3 className="text-xs font-bold text-gray-500 uppercase tracking-widest">{i18next.t("home:New users today")}</h3>
          <p className="text-3xl font-bold mt-1 text-gray-900">+{getNewUsersToday()}</p>
          <div className="w-full bg-gray-100 mt-4 h-1 rounded-full overflow-hidden">
            <div className="bg-emerald-500 h-full" style={{width: `${Math.min(getNewUsersToday() * 10, 100)}%`}}></div>
          </div>
        </div>

        <div className="bg-white p-6 rounded-xl border border-gray-200 shadow-sm transition-shadow hover:shadow-md">
          <div className="flex justify-between items-start mb-4">
            <span className="material-symbols-outlined text-blue-600 bg-blue-50 p-2 rounded-lg">organization</span>
          </div>
          <h3 className="text-xs font-bold text-gray-500 uppercase tracking-widest">{i18next.t("general:Organizations")}</h3>
          <p className="text-3xl font-bold mt-1 text-gray-900">{dashboardData.organizationCounts[30]}</p>
          <div className="w-full bg-gray-100 mt-4 h-1 rounded-full overflow-hidden">
            <div className="bg-blue-600 h-full" style={{width: "60%"}}></div>
          </div>
        </div>
      </div>
    );
  };

  if (dashboardData === null) {
    return (
      <div style={{display: "flex", justifyContent: "center", alignItems: "center", minHeight: "60vh"}}>
        <Spin size="large" tip={i18next.t("login:Loading")} style={{paddingTop: "10%"}} />
      </div>
    );
  }

  return (
    <div style={{display: "flex", justifyContent: "center", flexDirection: "column", alignItems: "center", width: "100%", padding: "24px"}}>
      <header style={{marginBottom: "32px", width: "100%", paddingLeft: "16px"}}>
        <h2 style={{fontSize: "28px", fontWeight: 800, letterSpacing: "-0.5px", color: "#191c1e", fontFamily: "Inter, sans-serif"}}>{i18next.t("home:Identity Overview") || "Identity Overview"}</h2>
        <p style={{fontSize: "14px", color: "#43474f", marginTop: "8px"}}>{i18next.t("home:Real-time status of the Casdoor authentication ecosystem.") || "Real-time status of the Casdoor authentication ecosystem."}</p>
      </header>

      {renderMetricCards()}

      <div style={{width: "100%", display: "grid", gridTemplateColumns: "1fr", gap: "32px", padding: "0 16px"}}>
        <section style={{backgroundColor: "#fff", padding: "32px", borderRadius: "12px", border: "1px solid #e6e8ea", boxShadow: "0 1px 3px rgba(0,0,0,0.06)"}}>
          <div style={{display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "24px"}}>
            <div>
              <h3 style={{fontSize: "18px", fontWeight: 700, color: "#191c1e"}}>{i18next.t("home:Past 30 Days")}</h3>
              <p style={{fontSize: "12px", color: "#43474f"}}>{i18next.t("home:Auth requests across all organizations") || "Auth requests across all organizations"}</p>
            </div>
          </div>
          <div ref={chartRef} id="echarts-chart" style={{width: "100%", height: "400px"}} />
        </section>
      </div>

      <Tour
        open={Setting.isMobile() ? false : isTourVisible}
        onClose={setIsTourToLocal}
        steps={getSteps()}
        indicatorsRender={(current, total) => (
          <span>
            {current + 1} / {total}
          </span>
        )}
        onFinish={handleTourComplete}
      />
    </div>
  );
};

export default Dashboard;
