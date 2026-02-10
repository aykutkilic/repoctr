package emoji

import "repoctr/pkg/models"

// Map returns an emoji for the given runtime type.
func Map(rt models.RuntimeType) string {
	switch rt {
	case models.RuntimeGo:
		return "🐹"
	case models.RuntimePython:
		return "🐍"
	case models.RuntimeJava:
		return "☕"
	case models.RuntimeTypeScript:
		return "🔷"
	case models.RuntimeJavaScript:
		return "🟡"
	case models.RuntimeDart:
		return "🎯"
	case models.RuntimeDotNet:
		return "🟣"
	case models.RuntimeRust:
		return "🦀"
	case models.RuntimeCpp:
		return "⚙️"
	default:
		return "📦"
	}
}
