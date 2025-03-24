//go:build integration

package destiny

import (
	"context"
	"fmt"
	"os"

	"github.com/taiidani/no-time-to-explain/internal/data"
)

func ExampleHelper_GetClanFish() {
	cache := data.NewCache()
	client := NewTokenClient(cache, os.Getenv("BUNGIE_API_KEY"))
	helper := NewHelper(client)

	def, metric, err := helper.GetClanFish(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	fmt.Println(def.DisplayProperties.Name)
	fmt.Println(metric.TotalFish)

	// Output: Total Fish Caught
	// 84410
}
