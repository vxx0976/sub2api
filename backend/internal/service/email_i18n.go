package service

// emailTexts holds translated strings for email templates.
type emailTexts struct {
	VerifySubject        string // "[{site}] Verification Code"
	VerifyTitle          string // "Your verification code is:"
	VerifyExpiry         string // "This code will expire in <strong>15 minutes</strong>."
	VerifyIgnore         string // "If you did not request this code, please ignore this email."
	AutoEmail            string // "This is an automated email, please do not reply."

	ResetSubject         string // "[{site}] Password Reset"
	ResetHeading         string // "Password Reset Request"
	ResetDescription     string // "You have requested to reset your password. Click the button below to set a new password:"
	ResetButton          string // "Reset Password"
	ResetExpiry          string // "This link will expire in <strong>30 minutes</strong>."
	ResetIgnore          string // "If you did not request a password reset, please ignore this email. Your password will remain unchanged."
	ResetLinkFallback    string // "If the button does not work, copy and paste the following link into your browser:"
}

var emailI18n = map[string]emailTexts{
	"zh": {
		VerifySubject:     "邮箱验证码",
		VerifyTitle:       "您的验证码是：",
		VerifyExpiry:      "此验证码将在 <strong>15 分钟</strong>后过期。",
		VerifyIgnore:      "如果您没有请求此验证码，请忽略此邮件。",
		AutoEmail:         "此邮件由系统自动发送，请勿回复。",
		ResetSubject:      "密码重置请求",
		ResetHeading:      "密码重置请求",
		ResetDescription:  "您已请求重置密码。请点击下方按钮设置新密码：",
		ResetButton:       "重置密码",
		ResetExpiry:       "此链接将在 <strong>30 分钟</strong>后失效。",
		ResetIgnore:       "如果您没有请求重置密码，请忽略此邮件。您的密码将保持不变。",
		ResetLinkFallback: "如果按钮无法点击，请复制以下链接到浏览器中打开：",
	},
	"zhTW": {
		VerifySubject:     "信箱驗證碼",
		VerifyTitle:       "您的驗證碼是：",
		VerifyExpiry:      "此驗證碼將在 <strong>15 分鐘</strong>後過期。",
		VerifyIgnore:      "如果您沒有請求此驗證碼，請忽略此郵件。",
		AutoEmail:         "此郵件由系統自動傳送，請勿回覆。",
		ResetSubject:      "密碼重置請求",
		ResetHeading:      "密碼重置請求",
		ResetDescription:  "您已請求重置密碼。請點擊下方按鈕設定新密碼：",
		ResetButton:       "重置密碼",
		ResetExpiry:       "此連結將在 <strong>30 分鐘</strong>後失效。",
		ResetIgnore:       "如果您沒有請求重置密碼，請忽略此郵件。您的密碼將保持不變。",
		ResetLinkFallback: "如果按鈕無法點擊，請複製以下連結到瀏覽器中開啟：",
	},
	"en": {
		VerifySubject:     "Verification Code",
		VerifyTitle:       "Your verification code is:",
		VerifyExpiry:      "This code will expire in <strong>15 minutes</strong>.",
		VerifyIgnore:      "If you did not request this code, please ignore this email.",
		AutoEmail:         "This is an automated email, please do not reply.",
		ResetSubject:      "Password Reset Request",
		ResetHeading:      "Password Reset Request",
		ResetDescription:  "You have requested to reset your password. Click the button below to set a new password:",
		ResetButton:       "Reset Password",
		ResetExpiry:       "This link will expire in <strong>30 minutes</strong>.",
		ResetIgnore:       "If you did not request a password reset, please ignore this email. Your password will remain unchanged.",
		ResetLinkFallback: "If the button does not work, copy and paste the following link into your browser:",
	},
	"ja": {
		VerifySubject:     "メール認証コード",
		VerifyTitle:       "認証コード：",
		VerifyExpiry:      "この認証コードは <strong>15 分間</strong>有効です。",
		VerifyIgnore:      "このコードをリクエストしていない場合は、このメールを無視してください。",
		AutoEmail:         "このメールはシステムにより自動送信されています。返信しないでください。",
		ResetSubject:      "パスワードリセット",
		ResetHeading:      "パスワードリセット",
		ResetDescription:  "パスワードのリセットがリクエストされました。下のボタンをクリックして新しいパスワードを設定してください：",
		ResetButton:       "パスワードをリセット",
		ResetExpiry:       "このリンクは <strong>30 分間</strong>有効です。",
		ResetIgnore:       "パスワードのリセットをリクエストしていない場合は、このメールを無視してください。パスワードは変更されません。",
		ResetLinkFallback: "ボタンが機能しない場合は、以下のリンクをコピーしてブラウザに貼り付けてください：",
	},
	"ko": {
		VerifySubject:     "이메일 인증 코드",
		VerifyTitle:       "인증 코드:",
		VerifyExpiry:      "이 인증 코드는 <strong>15분</strong> 후에 만료됩니다.",
		VerifyIgnore:      "이 코드를 요청하지 않으셨다면 이 이메일을 무시해 주세요.",
		AutoEmail:         "이 이메일은 자동으로 발송되었습니다. 회신하지 마세요.",
		ResetSubject:      "비밀번호 재설정 요청",
		ResetHeading:      "비밀번호 재설정 요청",
		ResetDescription:  "비밀번호 재설정이 요청되었습니다. 아래 버튼을 클릭하여 새 비밀번호를 설정해 주세요:",
		ResetButton:       "비밀번호 재설정",
		ResetExpiry:       "이 링크는 <strong>30분</strong> 후에 만료됩니다.",
		ResetIgnore:       "비밀번호 재설정을 요청하지 않으셨다면 이 이메일을 무시해 주세요. 비밀번호는 변경되지 않습니다.",
		ResetLinkFallback: "버튼이 작동하지 않으면 아래 링크를 복사하여 브라우저에 붙여넣으세요:",
	},
	"ru": {
		VerifySubject:     "Код подтверждения",
		VerifyTitle:       "Ваш код подтверждения:",
		VerifyExpiry:      "Код действителен в течение <strong>15 минут</strong>.",
		VerifyIgnore:      "Если вы не запрашивали этот код, проигнорируйте это письмо.",
		AutoEmail:         "Это автоматическое сообщение, не отвечайте на него.",
		ResetSubject:      "Сброс пароля",
		ResetHeading:      "Запрос на сброс пароля",
		ResetDescription:  "Вы запросили сброс пароля. Нажмите кнопку ниже, чтобы установить новый пароль:",
		ResetButton:       "Сбросить пароль",
		ResetExpiry:       "Эта ссылка действительна в течение <strong>30 минут</strong>.",
		ResetIgnore:       "Если вы не запрашивали сброс пароля, проигнорируйте это письмо. Ваш пароль останется прежним.",
		ResetLinkFallback: "Если кнопка не работает, скопируйте и вставьте следующую ссылку в браузер:",
	},
	"de": {
		VerifySubject:     "Bestätigungscode",
		VerifyTitle:       "Ihr Bestätigungscode:",
		VerifyExpiry:      "Dieser Code läuft in <strong>15 Minuten</strong> ab.",
		VerifyIgnore:      "Wenn Sie diesen Code nicht angefordert haben, ignorieren Sie diese E-Mail.",
		AutoEmail:         "Diese E-Mail wurde automatisch gesendet. Bitte antworten Sie nicht darauf.",
		ResetSubject:      "Passwort zurücksetzen",
		ResetHeading:      "Passwort zurücksetzen",
		ResetDescription:  "Sie haben das Zurücksetzen Ihres Passworts angefordert. Klicken Sie auf die Schaltfläche unten, um ein neues Passwort festzulegen:",
		ResetButton:       "Passwort zurücksetzen",
		ResetExpiry:       "Dieser Link läuft in <strong>30 Minuten</strong> ab.",
		ResetIgnore:       "Wenn Sie kein Zurücksetzen des Passworts angefordert haben, ignorieren Sie diese E-Mail. Ihr Passwort bleibt unverändert.",
		ResetLinkFallback: "Wenn die Schaltfläche nicht funktioniert, kopieren Sie den folgenden Link und fügen Sie ihn in Ihren Browser ein:",
	},
	"fr": {
		VerifySubject:     "Code de vérification",
		VerifyTitle:       "Votre code de vérification :",
		VerifyExpiry:      "Ce code expirera dans <strong>15 minutes</strong>.",
		VerifyIgnore:      "Si vous n'avez pas demandé ce code, veuillez ignorer cet e-mail.",
		AutoEmail:         "Cet e-mail a été envoyé automatiquement, veuillez ne pas y répondre.",
		ResetSubject:      "Réinitialisation du mot de passe",
		ResetHeading:      "Réinitialisation du mot de passe",
		ResetDescription:  "Vous avez demandé la réinitialisation de votre mot de passe. Cliquez sur le bouton ci-dessous pour définir un nouveau mot de passe :",
		ResetButton:       "Réinitialiser le mot de passe",
		ResetExpiry:       "Ce lien expirera dans <strong>30 minutes</strong>.",
		ResetIgnore:       "Si vous n'avez pas demandé de réinitialisation, veuillez ignorer cet e-mail. Votre mot de passe restera inchangé.",
		ResetLinkFallback: "Si le bouton ne fonctionne pas, copiez et collez le lien suivant dans votre navigateur :",
	},
	"es": {
		VerifySubject:     "Código de verificación",
		VerifyTitle:       "Su código de verificación es:",
		VerifyExpiry:      "Este código expirará en <strong>15 minutos</strong>.",
		VerifyIgnore:      "Si no solicitó este código, ignore este correo electrónico.",
		AutoEmail:         "Este correo fue enviado automáticamente, por favor no responda.",
		ResetSubject:      "Restablecimiento de contraseña",
		ResetHeading:      "Solicitud de restablecimiento de contraseña",
		ResetDescription:  "Ha solicitado restablecer su contraseña. Haga clic en el botón de abajo para establecer una nueva contraseña:",
		ResetButton:       "Restablecer contraseña",
		ResetExpiry:       "Este enlace expirará en <strong>30 minutos</strong>.",
		ResetIgnore:       "Si no solicitó un restablecimiento de contraseña, ignore este correo. Su contraseña no se modificará.",
		ResetLinkFallback: "Si el botón no funciona, copie y pegue el siguiente enlace en su navegador:",
	},
	"pt": {
		VerifySubject:     "Código de verificação",
		VerifyTitle:       "Seu código de verificação é:",
		VerifyExpiry:      "Este código expirará em <strong>15 minutos</strong>.",
		VerifyIgnore:      "Se você não solicitou este código, ignore este e-mail.",
		AutoEmail:         "Este e-mail foi enviado automaticamente, por favor não responda.",
		ResetSubject:      "Redefinição de senha",
		ResetHeading:      "Solicitação de redefinição de senha",
		ResetDescription:  "Você solicitou a redefinição da sua senha. Clique no botão abaixo para definir uma nova senha:",
		ResetButton:       "Redefinir senha",
		ResetExpiry:       "Este link expirará em <strong>30 minutos</strong>.",
		ResetIgnore:       "Se você não solicitou a redefinição de senha, ignore este e-mail. Sua senha permanecerá inalterada.",
		ResetLinkFallback: "Se o botão não funcionar, copie e cole o seguinte link no seu navegador:",
	},
}

// getEmailTexts returns translated email texts for the given locale, falling back to Chinese.
func getEmailTexts(locale string) emailTexts {
	if t, ok := emailI18n[locale]; ok {
		return t
	}
	// Try base language (e.g. "zh-CN" -> "zh")
	if len(locale) > 2 {
		base := locale[:2]
		if t, ok := emailI18n[base]; ok {
			return t
		}
	}
	return emailI18n["zh"]
}
