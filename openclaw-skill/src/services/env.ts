export function requireEnv(name: string): string {
  const value = process.env[name];
  if (!value) {
    throw new Error(`Missing required env: ${name}`);
  }
  return value;
}

export function getEnv(name: string, fallback: string): string {
  return process.env[name] ?? fallback;
}
