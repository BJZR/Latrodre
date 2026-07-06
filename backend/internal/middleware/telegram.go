package middleware

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"

	"latrode-fusion/internal/config"
	"latrode-fusion/internal/models"
)

func SendOrderNotification(cfg *config.Config, order *models.Order, user *models.User) {
	if cfg.TgBotToken == "" || cfg.TgChatID == "" {
		return
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("🛒 *Nueva Orden #%d*\n", order.ID))
	b.WriteString(fmt.Sprintf("👤 *Cliente:* %s\n", escapeMD(user.Username)))
	b.WriteString(fmt.Sprintf("📧 %s\n", escapeMD(user.Email)))
	b.WriteString(fmt.Sprintf("📞 %s\n", escapeMD(user.Phone)))
	b.WriteString(fmt.Sprintf("📍 %s, %s, %s %s\n",
		escapeMD(user.Address), escapeMD(user.City), escapeMD(user.PostalCode), escapeMD(user.Country)))
	b.WriteString(fmt.Sprintf("📦 *Envío a:* %s\n", escapeMD(order.ShippingName)))
	b.WriteString(fmt.Sprintf("💳 *Pago:* %s\n\n", escapeMD(order.PaymentMethod)))

	b.WriteString("┌── *Productos* ──\n")
	for i, item := range order.Items {
		line := fmt.Sprintf("│ %d. %s", i+1, escapeMD(item.ProductName))
		if item.ColorName != "" {
			line += fmt.Sprintf(" (%s)", escapeMD(item.ColorName))
		}
		if item.Size != "" {
			line += fmt.Sprintf(" [%s]", escapeMD(item.Size))
		}
		line += fmt.Sprintf(" x%d — $%.0f\n", item.Quantity, item.Subtotal)
		b.WriteString(line)
	}
	b.WriteString("└───────────────\n\n")

	b.WriteString(fmt.Sprintf("💰 *Total:* $%.0f", order.Total))

	go func() {
		body := fmt.Sprintf(`{"chat_id":"%s","text":"%s","parse_mode":"Markdown"}`,
			cfg.TgChatID, escapeJSON(b.String()))
		resp, err := http.Post(
			fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.TgBotToken),
			"application/json",
			bytes.NewReader([]byte(body)),
		)
		if err != nil {
			fmt.Printf("Telegram send error: %v\n", err)
			return
		}
		resp.Body.Close()
	}()
}

func escapeMD(s string) string {
	s = strings.ReplaceAll(s, "_", "\\_")
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "[", "\\[")
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}

func escapeJSON(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}
