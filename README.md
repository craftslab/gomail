# mailsender

[![Build Status](https://travis-ci.com/craftslab/mailsender.svg?branch=master)](https://travis-ci.com/craftslab/mailsender)
[![Code Coverage](http://gocover.io/_badge/github.com/craftslab/mailsender)](http://gocover.io/github.com/craftslab/mailsender)
[![License](https://img.shields.io/github/license/craftslab/mailsender.svg?color=brightgreen)](https://github.com/craftslab/mailsender/blob/master/LICENSE)
[![Tag](https://img.shields.io/github/tag/craftslab/mailsender.svg?color=brightgreen)](https://github.com/craftslab/mailsender/tags)



## Introduction

*Mail Sender* is a mail sender written in Go.



## Features

*Mail Sender* supports:
- Attachments
- HTML and text templates



## Examples

```
mailsender \
  --config "config/sender.json" \
  --attachment "attach1.txt,attach2.text" \
  --body "body.txt" \
  --content_type "PLAIN_TEXT" \
  --header "HEADER" \
  --recipients "alen@example.com,bob@example.com,cc:catherine@example.com" \
  --title "TITLE"

recipientparser \
  --config "config/parser.json" \
  --filter "@example1.com,@example2.com" \
  --recipients "alen,cc:bob@example.com"
```



## License

[Apache 2.0](LICENSE)
