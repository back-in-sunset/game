export function extractDemoJwt(accessToken: string): string {
  const [token] = accessToken.split("|");
  return token?.trim() ?? "";
}
