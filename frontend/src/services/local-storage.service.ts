export function setLocalBool(name: string, value: boolean) {
  localStorage.setItem(name, value ? 'true' : 'false');
}

export function getLocalBool(name: string): boolean {
  return localStorage.getItem(name) === 'true';
}
