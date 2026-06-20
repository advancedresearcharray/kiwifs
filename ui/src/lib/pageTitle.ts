import { titleize } from "./paths";

/** Browser tab title: page name plus app name, or app name alone on welcome. */
export function formatDocumentTitle(pagePath: string | null, appName: string): string {
  if (!pagePath) return appName;
  const pageTitle = titleize(pagePath);
  return pageTitle ? `${pageTitle} · ${appName}` : appName;
}
