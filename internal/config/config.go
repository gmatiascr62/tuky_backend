package config

import "os"

type Config struct {
	SupabaseURL     string
	SupabaseAnonKey string
	DatabaseURL     string
}

func Load() Config {
	return Config{
		SupabaseURL:     os.Getenv("SUPABASE_URL"),
		SupabaseAnonKey: os.Getenv("SUPABASE_ANON_KEY"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
	}
}
