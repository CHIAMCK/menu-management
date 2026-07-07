package messaging

import "log/slog"

func LogKitchenDisplayNotification(logger *slog.Logger, event OrderPlacedEvent) {
	if logger == nil {
		logger = slog.Default()
	}

	logger.Info("kitchen display notification",
		"event", "order.placed",
		"order_id", event.OrderID,
		"merchant_id", event.MerchantID,
		"user_id", event.UserID,
		"status", event.Status,
		"total_amount", event.TotalAmount,
		"items", event.Items,
		"created_at", event.CreatedAt,
	)
}
