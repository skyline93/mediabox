// src/services/axiosInstance.js
import axios from 'axios';
import { BASE_URL } from '../config';

// 创建 axios 实例
const axiosInstance = axios.create({
  baseURL: BASE_URL, // 替换为你的基础 URL
  timeout: 10000, // 可选: 设置请求超时（10秒）
});

// 添加请求拦截器
axiosInstance.interceptors.request.use(
  (config) => {
    // 从 localStorage 中获取 token
    const token = localStorage.getItem('token');
    if (token) {
      // 将 token 添加到 Authorization 头
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 添加响应拦截器
axiosInstance.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response && error.response.status === 401) {
      // 处理 token 过期的情况，比如提示用户重新登录
      alert('Session expired. Please log in again.');
      window.location.href = '/login'; // 或其他处理逻辑
    }
    return Promise.reject(error);
  }
);

export default axiosInstance;
