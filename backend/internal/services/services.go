package services

// Services holds all service instances
type Services struct {
	Auth    *AuthService
	Event   *EventService
	Payment *PaymentService
	Upload  *UploadService
	Visitor *VisitorService
}
