package types

import "time"

type Config struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Data      string    `json:"data"`
	CreatedAt time.Time `json:"created_at"`
}

type ConfigListOption struct {
}
