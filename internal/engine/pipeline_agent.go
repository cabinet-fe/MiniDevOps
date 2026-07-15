package engine

// Sync in-pipeline Agent stage is FORBIDDEN in 2.0 (DESIGN §5.2 / P1 Gate).
// Agent runs are created asynchronously on build events in P4
// (artifact_ready / distribution_finished). Do not reintroduce runMountedAgents.
