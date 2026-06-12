export type SeoMeta = {
  title: string
  description: string
  keywords?: string
  canonicalUrl: string
  imageUrl: string
  siteName: string
  locale: 'en' | 'zh'
  noindex?: boolean
  structuredData?: unknown | unknown[]
}

function upsertMeta(attributes: Record<string, string>, content: string) {
  const selector = `meta${Object.entries(attributes)
    .map(([key, value]) => `[${key}="${CSS.escape(value)}"]`)
    .join('')}`

  let tag = document.querySelector<HTMLMetaElement>(selector)
  if (!tag) {
    tag = document.createElement('meta')
    Object.entries(attributes).forEach(([key, value]) => tag!.setAttribute(key, value))
    document.head.appendChild(tag)
  }
  tag.setAttribute('content', content)
}

function upsertLink(rel: string, href: string) {
  let link = document.querySelector<HTMLLinkElement>(`link[rel="${CSS.escape(rel)}"]`)
  if (!link) {
    link = document.createElement('link')
    link.rel = rel
    document.head.appendChild(link)
  }
  link.href = href
}

function upsertAlternateLink(hreflang: string, href: string) {
  let link = document.querySelector<HTMLLinkElement>(
    `link[rel="alternate"][hreflang="${CSS.escape(hreflang)}"]`
  )
  if (!link) {
    link = document.createElement('link')
    link.rel = 'alternate'
    link.hreflang = hreflang
    document.head.appendChild(link)
  }
  link.href = href
}

function withLangParam(url: string, lang?: string) {
  try {
    const u = new URL(url)
    if (lang) {
      u.searchParams.set('lang', lang)
    } else {
      u.searchParams.delete('lang')
    }
    return u.toString()
  } catch {
    return url
  }
}

function upsertJsonLd(id: string, data: unknown) {
  let script = document.querySelector<HTMLScriptElement>(`script#${CSS.escape(id)}`)
  if (!script) {
    script = document.createElement('script')
    script.id = id
    script.type = 'application/ld+json'
    document.head.appendChild(script)
  }
  script.text = JSON.stringify(data)
}

function normalizeUrl(url: string) {
  try {
    return new URL(url).toString()
  } catch {
    return url
  }
}

export function applySeoMeta(meta: SeoMeta) {
  if (typeof window === 'undefined') return

  document.title = meta.title

  upsertMeta({ name: 'description' }, meta.description)
  if (meta.keywords) upsertMeta({ name: 'keywords' }, meta.keywords)
  upsertMeta({ name: 'robots' }, meta.noindex ? 'noindex,nofollow' : 'index,follow')

  upsertLink('canonical', meta.canonicalUrl)

  // index.html 静态写死的 hreflang/preconnect/dns-prefetch 指向主站，
  // 商户域名访问时必须跟随当前 origin 更新，否则会泄露主站域名
  upsertAlternateLink('zh', withLangParam(meta.canonicalUrl))
  upsertAlternateLink('en', withLangParam(meta.canonicalUrl, 'en'))
  upsertAlternateLink('ru', withLangParam(meta.canonicalUrl, 'ru'))
  upsertAlternateLink('x-default', withLangParam(meta.canonicalUrl))
  upsertLink('preconnect', window.location.origin)
  upsertLink('dns-prefetch', window.location.origin)

  // Open Graph
  upsertMeta({ property: 'og:type' }, 'website')
  upsertMeta({ property: 'og:site_name' }, meta.siteName)
  upsertMeta({ property: 'og:title' }, meta.title)
  upsertMeta({ property: 'og:description' }, meta.description)
  upsertMeta({ property: 'og:url' }, meta.canonicalUrl)
  upsertMeta({ property: 'og:image' }, meta.imageUrl)
  upsertMeta({ property: 'og:locale' }, meta.locale === 'zh' ? 'zh_CN' : 'en_US')

  // Twitter
  upsertMeta({ name: 'twitter:card' }, 'summary_large_image')
  upsertMeta({ name: 'twitter:title' }, meta.title)
  upsertMeta({ name: 'twitter:description' }, meta.description)
  upsertMeta({ name: 'twitter:image' }, meta.imageUrl)

  // Basic app hints
  upsertMeta({ name: 'application-name' }, meta.siteName)
  upsertMeta({ name: 'apple-mobile-web-app-title' }, meta.siteName)

  const origin = window.location.origin
  const webSite = {
    '@context': 'https://schema.org',
    '@type': 'WebSite',
    name: meta.siteName,
    url: origin,
    inLanguage: meta.locale === 'zh' ? 'zh-CN' : 'en',
    description: meta.description,
    image: normalizeUrl(meta.imageUrl)
  }

  const extra = meta.structuredData
  const jsonLd = (() => {
    if (!extra) return webSite
    const list = Array.isArray(extra) ? extra : [extra]
    return [webSite, ...list]
  })()

  upsertJsonLd('structured-data', jsonLd)
}
