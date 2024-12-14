"use client";
import { createContext, use, useEffect, useState } from "react";
import { isTokenExpired } from "./token";

type AuthProviderProps = {
  children: React.ReactNode;
};

type AuthProviderState = {
  token: string;
  email: string;
  isAuth: boolean;
  userID: number;
  setAuth: (token: string, email: string, userID: number) => void;
};

const AuthProviderContext = createContext<AuthProviderState>({
  token: "",
  email: "",
  userID: 0,
  isAuth: false,
  setAuth: () => null,
});

export function AuthProvider({ children }: AuthProviderProps) {
  const [email, setEmail] = useState(() => localStorage.getItem("email") || "");

  const [token, setToken] = useState(() => localStorage.getItem("token") || "");
  const [userID, setUserID] = useState(
    () => (localStorage.getItem("user_id") || 0) as number
  );

  useEffect(() => {
    localStorage.setItem("email", email);
    localStorage.setItem("token", token);
    localStorage.setItem("user_id", String(userID));
  }, [email, token]);

  const isAuth = !isTokenExpired(token) && userID !== 0 && email !== "";

  const value = {
    token: token,
    email: email,
    userID: userID,
    isAuth: isAuth,
    setAuth: (token: string, email: string, userID: number) => {
      setToken(token);
      setEmail(email);
      setUserID(userID);
    },
  };

  return <AuthProviderContext value={value}>{children}</AuthProviderContext>;
}

export const useAuth = () => {
  const context = use(AuthProviderContext);

  if (context === undefined) {
    throw new Error("useAuth must be use within a AuthProvider");
  }

  return context;
};
