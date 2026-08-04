package mytools

import (
	"bytes"
	"fmt"
	"html/template"
	"log"

	"gopkg.in/gomail.v2"
)

// EmailConfig 邮件服务器配置结构体（工程化标准写法：方便从 yaml/json 配置文件读取）
type EmailConfig struct {
	Host     string // SMTP 服务器地址，如 "smtp.163.com"
	Port     int    // 端口，通常为 465 (SSL) 或 587
	Username string // 发件人邮箱，如 "xxxx@163.com"
	Password string // 16位授权码（非网页登录密码）
	FromTitle string // 发件人显示名称，如 "我的Go程序助手"
}

// SendMail 发送邮件核心函数
// toMails: 收件人列表
// subject: 邮件主题
// content: 邮件正文（支持纯文本或带 HTML 标签的内容）
// isHTML: 是否渲染为 HTML 网页格式
func SendMail(cfg EmailConfig, toMails []string, subject string, content string, isHTML bool) error {
	m := gomail.NewMessage()

	// 1. 设置邮件头信息
	m.SetHeader("From", m.FormatAddress(cfg.Username, cfg.FromTitle)) // 发件人：展示名 <邮箱地址>
	m.SetHeader("To", toMails...)                                      // 收件人（支持一次发给多人）
	m.SetHeader("Subject", subject)                                    // 主题

	// 2. 根据类型渲染正文内容
	if isHTML {
		// 使用 HTML 模板进行安全渲染
		bodyHTML, err := renderHTMLTemplate(subject, content)
		if err != nil {
			return fmt.Errorf("渲染 HTML 邮件模板失败: %w", err)
		}
		m.SetBody("text/html", bodyHTML)
	} else {
		// 普通纯文本格式
		m.SetBody("text/plain", content)
	}

	// 3. 构建 SMTP 拨号器
	d := gomail.NewDialer(cfg.Host, cfg.Port, cfg.Username, cfg.Password)

	// 如果使用的是 465 端口且遇到 TLS 证书校验报错，可取消下面这行的注释跳过验证：
	// d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// 4. 执行发送
	if err := d.DialAndSend(m); err != nil {
		log.Printf("[ERROR] 邮件发送失败: host=%s, port=%d, err=%v\n", cfg.Host, cfg.Port, err)
		return err
	}

	log.Printf("[INFO] 邮件成功发送给 %v\n", toMails)
	return nil
}

// 内部函数：基于 html/template 安全渲染漂亮的 HTML 邮件页面
func renderHTMLTemplate(title, content string) (string, error) {
	// 优雅、高兼容性的邮件 HTML 模版
	const tplStr = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <title>{{.Title}}</title>
</head>
<body style="margin: 0; padding: 0; background-color: #f4f6f9; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;">
    <table role="presentation" width="100%" cellspacing="0" cellpadding="0" style="background-color: #f4f6f9; padding: 40px 0;">
        <tr>
            <td align="center">
                <table role="presentation" width="580" cellspacing="0" cellpadding="0" style="background: #ffffff; border-radius: 8px; overflow: hidden; box-shadow: 0 4px 12px rgba(0,0,0,0.05);">
                    <!-- 头部 -->
                    <tr>
                        <td style="background-color: #2563eb; padding: 24px; text-align: center;">
                            <h1 style="color: #ffffff; margin: 0; font-size: 20px;">{{.Title}}</h1>
                        </td>
                    </tr>
                    <!-- 内容区 -->
                    <tr>
                        <td style="padding: 32px 24px; color: #334155; font-size: 15px; line-height: 1.6;">
                            {{.Content}}
                        </td>
                    </tr>
                    <!-- 页脚 -->
                    <tr>
                        <td style="background-color: #f8fafc; padding: 16px; text-align: center; color: #94a3b8; font-size: 12px; border-top: 1px solid #e2e8f0;">
                            此邮件由 Go 系统自动发送，请勿直接回复。
                        </td>
                    </tr>
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`

	tmpl, err := template.New("email").Parse(tplStr)
	if err != nil {
		return "", err
	}

	data := struct {
		Title   string
		Content template.HTML // 允许传入 HTML 标签（如果想纯文本，换成 string 即可自动转义）
	}{
		Title:   title,
		Content: template.HTML(content),
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}