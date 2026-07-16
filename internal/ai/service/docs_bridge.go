package service

// DocsBridge adapts AgentService for the project docs/generate hook.
type DocsBridge struct {
	agents *AgentService
}

func NewDocsBridge(agents *AgentService) *DocsBridge {
	return &DocsBridge{agents: agents}
}

func (b *DocsBridge) StartDocsGenerate(userID, projectID, nodeID, agentID uint) (uint, error) {
	run, err := b.agents.DocsGenerateRun(agentID, userID, projectID, nodeID)
	if err != nil {
		return 0, err
	}
	return run.ID, nil
}
