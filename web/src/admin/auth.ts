const ADMIN_TOKEN_KEY = 'cyber_range_admin_token';
const ADMIN_INFO_KEY = 'cyber_range_admin_info';

export interface AdminInfo {
  id: string;
  username: string;
  email: string;
  name: string;
}

export function isAdminAuthed(): boolean {
  return !!localStorage.getItem(ADMIN_TOKEN_KEY);
}

export function setAdminAuth(token: string, adminInfo: AdminInfo): void {
  localStorage.setItem(ADMIN_TOKEN_KEY, token);
  localStorage.setItem(ADMIN_INFO_KEY, JSON.stringify(adminInfo));
}

export function getAdminToken(): string | null {
  return localStorage.getItem(ADMIN_TOKEN_KEY);
}

export function getAdminInfo(): AdminInfo | null {
  const info = localStorage.getItem(ADMIN_INFO_KEY);
  return info ? JSON.parse(info) : null;
}

export function adminLogout(): void {
  localStorage.removeItem(ADMIN_TOKEN_KEY);
  localStorage.removeItem(ADMIN_INFO_KEY);
}
