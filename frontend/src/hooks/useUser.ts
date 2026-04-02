import { useState, useEffect } from 'react';
import { getUser, updateUser } from '../api/user';
import type { User } from '../api/user';
import { getLocalUserId } from '../utils/user';

export const useUser = () => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUser = async () => {
      try {
        const userId = getLocalUserId();
        const userData = await getUser(userId);
        setUser(userData);
      } catch (err: any) {
        setError(err.message || 'Failed to fetch user');
      } finally {
        setLoading(false);
      }
    };

    fetchUser();
  }, []);

  const updateUserInfo = async (nickname?: string, avatar?: string) => {
    if (!user) return;
    try {
      const updatedUser = await updateUser(user.id, { nickname, avatar });
      setUser(updatedUser);
      return updatedUser;
    } catch (err: any) {
      throw new Error(err.message || 'Failed to update user');
    }
  };

  return { user, loading, error, updateUserInfo };
};
