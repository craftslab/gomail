<div align="center">

# gomail

[![Actions Status](https://github.com/craftslab/gomail/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/craftslab/gomail/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/craftslab/gomail)](https://goreportcard.com/report/github.com/craftslab/gomail)
[![License](https://img.shields.io/github/license/craftslab/gomail.svg?color=brightgreen)](https://github.com/craftslab/gomail/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/gomail.svg?color=brightgreen)](https://github.com/craftslab/gomail/tags)

**一个功能强大且灵活的 Go 语言邮件发送工具**

[English](README.md) | [简体中文](README_cn.md)

</div>

---

## 📖 简介

**gomail** 是一个用 Go 语言编写的强大邮件发送工具，旨在简化邮件投递，支持附件、模板和灵活的收件人管理。

## ⚙️ 前置要求

- **Go** >= 1.24.0

## ✨ 功能特性

**gomail** 提供全面的邮件功能：

- 📎 **附件支持** - 轻松发送多个文件附件
- 📝 **HTML 和文本模板** - 支持 HTML 和纯文本内容
- 👥 **收件人管理** - 高级收件人解析，支持抄送
- 🔍 **过滤功能** - 邮件域名过滤能力
- 🧪 **试运行模式** - 验证收件人而不实际发送邮件

## 🚀 快速开始

### 从源码构建

```bash
# 克隆仓库
git clone https://github.com/craftslab/gomail.git

# 进入项目目录
cd gomail

# 构建项目
make build
```

编译后的二进制文件将生成在 `bin/` 目录下。

## 📋 使用方法

### 解析器工具

解析和过滤收件人邮箱地址：

```bash
./parser \
  --config="config/parser.json" \
  --filter="@example1.com,@example2.com" \
  --recipients="alen,cc:bob@example.com"
```

### 发送器工具

使用各种选项发送邮件：

```bash
./sender \
  --config="config/sender.json" \
  --attachment="attach1.txt,attach2.txt" \
  --body="body.txt" \
  --content_type="PLAIN_TEXT" \
  --header="HEADER" \
  --recipients="alen@example.com,bob@example.com,cc:catherine@example.com" \
  --title="TITLE"
```

## 📚 命令行参考

### 解析器命令

**描述：** 解析和过滤邮件收件人

```bash
usage: parser --recipients=RECIPIENTS [<flags>]

收件人解析器

标志参数:
      --help                   显示上下文相关的帮助信息（也可尝试 --help-long
                               和 --help-man）
      --version                显示应用程序版本
  -c, --config=CONFIG          配置文件，格式：.json
  -f, --filter=FILTER          过滤列表，格式：@example1.com,@example2.com
  -r, --recipients=RECIPIENTS  收件人列表，格式：alen,cc:bob@example.com
```

### 发送器命令

**描述：** 发送带有附件和模板的邮件

```bash
usage: sender --recipients=RECIPIENTS [<flags>]

邮件发送器

标志参数:
      --help                     显示上下文相关的帮助信息（也可尝试
                                 --help-long 和 --help-man）
      --version                  显示应用程序版本
  -a, --attachment=ATTACHMENT    附件文件，格式：attach1,attach2,...
  -b, --body=BODY                正文文本或文件
  -c, --config=CONFIG            配置文件，格式：.json
  -e, --content_type=PLAIN_TEXT  内容类型，格式：HTML 或 PLAIN_TEXT（默认）
  -r, --header=HEADER            头部文本
  -p, --recipients=RECIPIENTS    收件人列表，格式：
                                 alen@example.com,cc:bob@example.com
  -t, --title=TITLE              标题文本
  -n, --dry-run                  仅输出收件人验证 JSON 并退出；
                                 不实际发送邮件
```

## 📄 许可证

本项目采用 [LICENSE](LICENSE) 文件中规定的条款进行许可。

## 🔗 相关项目

- [rsmail](https://github.com/craftslab/rsmail) - 相关邮件项目

---

<div align="center">

由 [craftslab](https://github.com/craftslab) 用 ❤️ 制作

</div>
