export type Router = {
	isProtected(path: string): boolean
	redirectFor(path: string): string | null
	canAccess(path: string): boolean
}

const isProtectedPath = (path: string): boolean => {
	return path === '/guests' || path.startsWith('/guests/')
}

export function createRouter(opts: { token: string | null }): Router {
	const isProtected = (path: string): boolean => {
		return isProtectedPath(path)
	}

	const redirectFor = (path: string): string | null => {
		if (isProtectedPath(path) && !opts.token) {
			return '/login'
		}
		return null
	}

	const canAccess = (path: string): boolean => {
		if (!isProtectedPath(path)) {
			return true
		}
		return !!opts.token
	}

	return { isProtected, redirectFor, canAccess }
}
