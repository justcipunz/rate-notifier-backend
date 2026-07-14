package app

const (
	messageMethodNotAllowed      = "Метод не поддерживается"
	messageInvalidRequestData    = "Некорректные данные запроса"
	messageInvalidEmail          = "Некорректный email"
	messageInvalidCredentials    = "Неверный email или пароль"
	messageAuthRequired          = "Требуется авторизация"
	messageInternalError         = "Внутренняя ошибка"
	messageEmailAlreadyExists    = "Email уже зарегистрирован"
	messagePasswordTooShort      = "Пароль должен содержать не менее 8 символов"
	messagePasswordTooLong       = "Пароль должен содержать не более 72 байт"
	messageInvalidTargetID       = "Некорректный ID цели"
	messageTargetNotFound        = "Цель не найдена"
	messageInvalidNotificationID = "Некорректный ID уведомления"
	messageNotificationNotFound  = "Уведомление не найдено"
	messageSettingsNotFound      = "Настройки не найдены"
	messageNotificationsRequired = "Поле notifications_enabled обязательно"
	messageIsActiveRequired      = "Поле is_active обязательно"
	messageCurrencyNotSupported  = "Валюта не поддерживается"
	messageTargetValuePositive   = "Значение цели должно быть больше нуля"
	messageUnknownCondition      = "Недопустимое условие"
)
