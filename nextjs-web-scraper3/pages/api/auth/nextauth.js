// pages/api/auth/[...nextauth].js
import NextAuth from 'next-auth';
import GitHubProvider from 'next-auth/providers/github';

export default NextAuth({
  providers: [
    // OAuth authentication providers...
    AppleProvider({
        clientId: process.env.APPLE_ID,
        clientSecret: process.env.APPLE_SECRET
      }),
      FacebookProvider({
        clientId: process.env.FACEBOOK_ID,
        clientSecret: process.env.FACEBOOK_SECRET
      }),
      GoogleProvider({
        clientId: process.env.GOOGLE_ID,
        clientSecret: process.env.GOOGLE_SECRET
      }),
      // Passwordless / email sign in
      EmailProvider({
        server: process.env.MAIL_SERVER,
        from: 'NextAuth.js <no-reply@example.com>'
      }),
  ],
  pages: {
    signIn: '/auth/signin',
  },
});
