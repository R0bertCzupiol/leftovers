package logicalrouting

type logger interface {
	Printf(message string, a ...interface{})
	PromptWithDetails(resourceType, resourceName string) bool
}
