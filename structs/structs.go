package structs

type Node struct {
	Name      string `json:"name"`
	Partition string `json:"partition"`
	Address   string `json:"address"`
	Labels    struct {
		NodeRoleKubernetesIoWorker string `json:"node-role.kubernetes.io/worker"`
	}
}

type Members struct {
	Kind     string `json:"kind"`
	SelfLink string `json:"selfLink"`
	Items    []struct {
		Kind            string `json:"kind"`
		Name            string `json:"name"`
		Partition       string `json:"partition"`
		FullPath        string `json:"fullPath"`
		Generation      int    `json:"generation"`
		SelfLink        string `json:"selfLink"`
		Address         string `json:"address"`
		ConnectionLimit int    `json:"connectionLimit"`
		DynamicRatio    int    `json:"dynamicRatio"`
		Ephemeral       string `json:"ephemeral"`
		Fqdn            struct {
			Autopopulate string `json:"autopopulate"`
		} `json:"fqdn"`
		InheritProfile string `json:"inheritProfile"`
		Logging        string `json:"logging"`
		Monitor        string `json:"monitor"`
		PriorityGroup  int    `json:"priorityGroup"`
		RateLimit      string `json:"rateLimit"`
		Ratio          int    `json:"ratio"`
		Session        string `json:"session"`
		State          string `json:"state"`
	} `json:"items"`
}

type Memberlist struct {
	Members []Memberlistmember `json:"members"`
}

type Memberlistmember struct {
	Name    string `json:"name"`
	Monitor string `json:"monitor"`
}
