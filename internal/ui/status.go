package ui

import (
	"fmt"
	"strings"
	"time"
)

type StatusInfo struct {
	Connected    bool
	Protocol     string
	ServerName   string
	Endpoint     string
	TunnelIP     string
	ConnectedAt  time.Time
	TxBytes      int64
	RxBytes      int64
	LastHandshake time.Time
}

func RenderStatus(s StatusInfo) string {
	var sb strings.Builder

	sb.WriteString(Banner())
	sb.WriteString("\n")

	if !s.Connected {
		sb.WriteString(BoxStyle.Render(
			WarningStyle.Render("● Disconnected") + "\n\n" +
			DimStyle.Render("Run 'voidvpn connect <server>' to connect."),
		))
		sb.WriteString("\n")
		return sb.String()
	}

	uptime := time.Since(s.ConnectedAt).Truncate(time.Second)

	protoLabel := "WireGuard"
	if s.Protocol == "openvpn" {
		protoLabel = "OpenVPN"
	}

	content := fmt.Sprintf(
		"%s\n\n%s %s\n%s %s\n%s %s\n%s %s\n%s %s\n%s %s\n%s %s / %s",
		SuccessStyle.Render("● Connected"),
		LabelStyle.Render("Server:"),
		ValueStyle.Render(s.ServerName),
		LabelStyle.Render("Protocol:"),
		ValueStyle.Render(protoLabel),
		LabelStyle.Render("Endpoint:"),
		ValueStyle.Render(s.Endpoint),
		LabelStyle.Render("Tunnel IP:"),
		ValueStyle.Render(s.TunnelIP),
		LabelStyle.Render("Uptime:"),
		AccentStyle.Render(uptime.String()),
		LabelStyle.Render("Last Handshake:"),
		ValueStyle.Render(formatHandshake(s.LastHandshake)),
		LabelStyle.Render("Traffic:"),
		AccentStyle.Render("↑ "+FormatBytes(s.TxBytes)),
		AccentStyle.Render("↓ "+FormatBytes(s.RxBytes)),
	)

	sb.WriteString(BoxStyle.Render(content))
	sb.WriteString("\n")
	return sb.String()
}

func formatHandshake(t time.Time) string {
	if t.IsZero() {
		return "never"
	}
	ago := time.Since(t).Truncate(time.Second)
	return fmt.Sprintf("%s ago", ago)
}
