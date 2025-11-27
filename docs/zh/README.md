# AINAR 文档

欢迎使用 AINAR 文档。AINAR（Ainar Is Not an AI Router）是一个统一的 OpenAI 兼容 API 网关，可以将请求路由到多个 AI 提供商。

> **注意**：此文档可能已过时。请参阅[英文文档](../en/README.md)获取最新信息。

## 目录

- [快速开始](getting-started.md)
- [安装指南](installation.md)
- [配置参考](configuration.md)
- [CLI 参考](cli.md)
- [WebUI 指南](webui.md)
- [API 参考](api.md)

## 概述

AINAR 提供单一的 OpenAI 兼容端点，可以根据模型名称前缀将请求路由到多个 AI 提供商。

### 工作原理

1. 您的应用程序向 AINAR 发送带有模型名称（如 `openai/gpt-4`）的请求
2. AINAR 提取提供商代码（`openai`）和模型名称（`gpt-4`）
3. AINAR 将请求路由到相应的提供商
4. 响应以 OpenAI 兼容格式返回

## 快速链接

- [GitHub 仓库](https://github.com/HeavySnowJakarta/ainar)
- [发布版本](https://github.com/HeavySnowJakarta/ainar/releases)
- [问题追踪](https://github.com/HeavySnowJakarta/ainar/issues)
