import React from 'react';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Card } from '../../common';

const MetricChart = ({ title, data, dataKey, color, name, domain = [0, 100] }) => (
  <Card title={title} className="chart-card">
    <ResponsiveContainer width="100%" height={250}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="time" />
        <YAxis domain={domain} />
        <Tooltip />
        <Legend />
        <Line 
          type="monotone" 
          dataKey={dataKey} 
          stroke={color} 
          name={name} 
          strokeWidth={2} 
          dot={false} 
        />
      </LineChart>
    </ResponsiveContainer>
  </Card>
);

export const CpuChart = ({ data }) => (
  <MetricChart
    title="CPU 使用率 (%)"
    data={data}
    dataKey="usage"
    color="#8884d8"
    name="CPU使用率"
    domain={[0, 100]}
  />
);

export const MemoryChart = ({ data }) => (
  <MetricChart
    title="内存使用率 (%)"
    data={data}
    dataKey="usage"
    color="#82ca9d"
    name="内存使用率"
    domain={[0, 100]}
  />
);

export const NetworkChart = ({ data }) => (
  <Card title="网络流量 (MB)" className="chart-card">
    <ResponsiveContainer width="100%" height={250}>
      <LineChart data={data}>
        <CartesianGrid strokeDasharray="3 3" />
        <XAxis dataKey="time" />
        <YAxis />
        <Tooltip />
        <Legend />
        <Line type="monotone" dataKey="rx" stroke="#8884d8" name="接收" strokeWidth={2} dot={false} />
        <Line type="monotone" dataKey="tx" stroke="#82ca9d" name="发送" strokeWidth={2} dot={false} />
      </LineChart>
    </ResponsiveContainer>
  </Card>
);

export default MetricChart;
