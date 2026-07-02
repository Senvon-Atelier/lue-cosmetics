import { createContext, useContext, ReactNode } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import {
  postAuthSignup,
  postAuthLogin,
  postAuthLogout,
} from '../api/generated/rueCosmeticsAPI';
import { sessionQueryOptions, Session } from './session';

interface AuthContextType {
  user: Session | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
  signup: (email: string, password: string, name?: string) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const queryClient = useQueryClient();
  const { data: user = null, isLoading } = useQuery(sessionQueryOptions);

  const refreshSession = async () => {
    await queryClient.invalidateQueries({ queryKey: sessionQueryOptions.queryKey });
  };

  const signup = async (email: string, password: string, name?: string) => {
    await postAuthSignup({ email, password, name });
    await refreshSession();
  };

  const login = async (email: string, password: string) => {
    await postAuthLogin({ email, password });
    await refreshSession();
  };

  const logout = async () => {
    await postAuthLogout();
    queryClient.setQueryData(sessionQueryOptions.queryKey, null);
    await queryClient.invalidateQueries();
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isLoading,
        isAuthenticated: !!user,
        isAdmin: user?.role === 'admin',
        signup,
        login,
        logout,
        refreshSession,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
