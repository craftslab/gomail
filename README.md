<div align="center">

# gomail

[![Actions Status](https://github.com/craftslab/gomail/workflows/ci/badge.svg?branch=main&event=push)](https://github.com/craftslab/gomail/actions?query=workflow%3Aci)
[![Go Report Card](https://goreportcard.com/badge/github.com/craftslab/gomail)](https://goreportcard.com/report/github.com/craftslab/gomail)
[![License](https://img.shields.io/github/license/craftslab/gomail.svg?color=brightgreen)](https://github.com/craftslab/gomail/blob/main/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/gomail.svg?color=brightgreen)](https://github.com/craftslab/gomail/tags)

**A powerful and flexible mail sender written in Go**

[English](README.md) | [简体中文](README_cn.md)

</div>

---

## 📖 Introduction

**gomail** is a robust mail sending utility written in Go, designed to simplify email delivery with support for attachments, templates, and flexible recipient management.

## ⚙️ Prerequisites

- **Go** >= 1.24.0

## ✨ Features

**gomail** provides comprehensive email functionality:

- 📎 **Attachments** - Send multiple file attachments with ease
- 📝 **HTML and Text Templates** - Support for both HTML and plain text content
- 👥 **Recipient Management** - Advanced recipient parsing with CC support
- 🔍 **Filtering** - Email domain filtering capabilities
- 🧪 **Dry Run Mode** - Validate recipients without sending

## 🚀 Quick Start

### Build from Source

```bash
# Clone the repository
git clone https://github.com/craftslab/gomail.git

# Navigate to the project directory
cd gomail

# Build the project
make build
```

The compiled binaries will be available in the `bin/` directory.

## 📋 Usage

### Parser Tool

Parse and filter recipient email addresses:

```bash
./parser \
  --config="config/parser.json" \
  --filter="@example1.com,@example2.com" \
  --recipients="alen,cc:bob@example.com"
```

### Sender Tool

Send emails with various options:

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

## 📚 Command Line Reference

### Parser Command

**Description:** Parse and filter email recipients

```bash
usage: parser --recipients=RECIPIENTS [<flags>]

Recipient parser

Flags:
      --help                   Show context-sensitive help (also try --help-long
                               and --help-man).
      --version                Show application version.
  -c, --config=CONFIG          Config file, format: .json
  -f, --filter=FILTER          Filter list, format: @example1.com,@example2.com
  -r, --recipients=RECIPIENTS  Recipients list, format: alen,cc:bob@example.com
```

### Sender Command

**Description:** Send emails with attachments and templates

```bash
usage: sender --recipients=RECIPIENTS [<flags>]

Mail sender

Flags:
      --help                     Show context-sensitive help (also try
                                 --help-long and --help-man).
      --version                  Show application version.
  -a, --attachment=ATTACHMENT    Attachment files, format: attach1,attach2,...
  -b, --body=BODY                Body text or file
  -c, --config=CONFIG            Config file, format: .json
  -e, --content_type=PLAIN_TEXT  Content type, format: HTML or PLAIN_TEXT
                                 (default)
  -r, --header=HEADER            Header text
  -p, --recipients=RECIPIENTS    Recipients list, format:
                                 alen@example.com,cc:bob@example.com
  -t, --title=TITLE              Title text
  -n, --dry-run                  Only output recipient validation JSON and exit;
                                 do not send
```

## 📄 License

This project is licensed under the terms specified in the [LICENSE](LICENSE) file.

## 🔗 Related Projects

- [rsmail](https://github.com/craftslab/rsmail) - Related mail project

---

<div align="center">

Made with ❤️ by [craftslab](https://github.com/craftslab)

</div>
