import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import {
  postAuthSignup,
  postAuthLogin,
  postAuthLogout,
  getAuthSession,
} from '../api/generated/rueCosmeticsAPI';

// Auth types
interface User {
  user_id: string;
  email: string;
  name?: string;
  email_verified: boolean;
  role: 'customer' | 'admin';
}

interface AuthState {
  user: User | null;
  isLoading: boolean;
  isAuthenticated: boolean;
  isAdmin: boolean;
}

interface AuthContextType extends AuthState {
  signup: (email: string, password: string, name?: string) => Promise<void>;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshSession: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

// Auth provider component
export function AuthProvider({ children }: { children: ReactNode }) {
  const [state, setState] = useState<AuthState>({
    user: null,
    isLoading: true,
    isAuthenticated: false,
    isAdmin: false,
  });

  // Refresh session from backend
  const refreshSession = async () => {
    try {
      const user = await getAuthSession<User>();
      setState({
        user,
        isLoading: false,
        isAuthenticated: !!user,
        isAdmin: user?.role === 'admin',
      });
    } catch (error) {
      setState({
        user: null,
        isLoading: false,
        isAuthenticated: false,
        isAdmin: false,
      });
    }
  };

  // Load session on mount
  useEffect(() => {
    refreshSession();
  }, []);

  // Signup
  const signup = async (email: string, password: string, name?: string) => {
    await postAuthSignup({ email, password, name });
    await refreshSession();
  };

  // Login
  const login = async (email: string, password: string) => {
    await postAuthLogin({ email, password });
    await refreshSession();
  };

  // Logout
  const logout = async () => {
    await postAuthLogout();
    setState({
      user: null,
      isLoading: false,
      isAuthenticated: false,
      isAdmin: false,
    });
  };

  return (
    <AuthContext.Provider
      value={{
        ...state,
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

// Hook to use auth context
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
