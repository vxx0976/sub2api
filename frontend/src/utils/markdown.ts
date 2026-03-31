import { Marked, Renderer } from 'marked'
import DOMPurify from 'dompurify'

const renderer = new Renderer()
const originalLink = renderer.link.bind(renderer)
renderer.link = function (token) {
  const html = originalLink(token)
  // Add target="_blank" and rel="noopener noreferrer" to all links
  return html.replace('<a ', '<a target="_blank" rel="noopener noreferrer" ')
}

const md = new Marked({ breaks: true, gfm: true, renderer })

export function renderMarkdown(content: string): string {
  if (!content) return ''
  return DOMPurify.sanitize(md.parse(content) as string, {
    ADD_ATTR: ['target'],
  })
}
