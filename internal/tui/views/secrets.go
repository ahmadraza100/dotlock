package views

func SecretsView(revealed bool) string {
	mask := "••••••••"
	if revealed {
		mask = "secret-value"
	}
	return "\n DATABASE_URL    " + mask + "\n REDIS_URL       " + mask + "\n"
}
