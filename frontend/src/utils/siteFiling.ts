export interface SiteFilingConfig {
  ICP_BEIAN_NUMBER?: string
  PUBLIC_SECURITY_BEIAN_NUMBER?: string
  PUBLIC_SECURITY_BEIAN_URL?: string
  PUBLIC_SECURITY_BEIAN_ICON_URL?: string
}

function safeAddress(value: string, allowLocal: boolean): boolean {
  if (!value || /[\s\\]/u.test(value)) return false
  if (allowLocal && value.startsWith('/') && !value.startsWith('//')) return true
  try {
    const url = new URL(value)
    return url.protocol === 'https:' && !!url.hostname && !url.username && !url.password
  } catch {
    return false
  }
}

export function resolveSiteFiling(config: SiteFilingConfig = {}) {
  const icp = config.ICP_BEIAN_NUMBER?.trim() || ''
  const number = config.PUBLIC_SECURITY_BEIAN_NUMBER?.trim() || ''
  const url = config.PUBLIC_SECURITY_BEIAN_URL?.trim() || ''
  const icon = config.PUBLIC_SECURITY_BEIAN_ICON_URL?.trim() || ''
  const police = number && safeAddress(url, false) && safeAddress(icon, true)
    ? { number, url, icon } : null
  return { icp, police, visible: !!(icp || police) }
}
