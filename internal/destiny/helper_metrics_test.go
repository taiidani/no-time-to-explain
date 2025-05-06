//go:build integration

package destiny

// func ExampleHelper_GetClanFish() {
// 	cache := data.NewCache()
// 	client := NewTokenClient(cache, os.Getenv("BUNGIE_API_KEY"))
// 	helper := NewHelper(client)

// 	def, metric, err := helper.GetClanFish(context.Background())
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		return
// 	}

// 	fmt.Println(def.DisplayProperties.Name)
// 	fmt.Println(metric.TotalFish)

// 	// Output: Total Fish Caught
// 	// 84410
// }

// func ExampleHelper_getTitles() {
// 	cache := data.NewCache()
// 	client := NewTokenClient(cache, os.Getenv("BUNGIE_API_KEY"))
// 	helper := NewHelper(client)

// 	manifest, err := client.GetManifest(context.Background())
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		return
// 	}

// 	titles, err := helper.getTitles(context.Background(), *manifest)
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		return
// 	}

// 	for _, title := range titles {
// 		_ = json.NewEncoder(os.Stdout).Encode(title)
// 	}

// 	// Output: Uh oh
// 	// 84410
// }

// func ExampleHelper_getPlayerTitleInfo() {
// 	const taiidaniMembershipType = 3
// 	const taiidaniMembershipID = "4611686018467493133"

// 	cache := data.NewCache()
// 	client := NewTokenClient(cache, os.Getenv("BUNGIE_API_KEY"))
// 	helper := NewHelper(client)

// 	manifest, err := client.GetManifest(context.Background())
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		return
// 	}

// 	titles, err := helper.getTitles(context.Background(), *manifest)
// 	if err != nil {
// 		fmt.Fprintln(os.Stderr, err)
// 		return
// 	}

// 	info, err := helper.getPlayerTitleInfo(context.Background(), titles, ClanMember{
// 		MemberType: taiidaniMembershipType,
// 		DestinyUserInfo: ClanMemberInfo{
// 			MembershipType: taiidaniMembershipType,
// 			MembershipID:   taiidaniMembershipID,
// 		},
// 	})

// 	_ = json.NewEncoder(os.Stdout).Encode(info)
// 	// Output: Uh oh
// }
