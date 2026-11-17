package postgres

import (
	"os"
	"uuid"
)

func testUUID(name string) (uuid.UUID, error) {
	return uuid.Parse(os.Getenv(name))
}

func testDifferentUUID(value uuid.UUID) uuid.UUID {
	other := uuid.New()
	for other == value {
		other = uuid.New()
	}
	return other
}
