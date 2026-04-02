export const getLocalUserId = (): string => {
  let userId = localStorage.getItem('texas_user_id');
  if (!userId) {
    // Generate a simple random ID for the user
    userId = 'user_' + Math.random().toString(36).substring(2, 10);
    localStorage.setItem('texas_user_id', userId);
  }
  return userId;
};
