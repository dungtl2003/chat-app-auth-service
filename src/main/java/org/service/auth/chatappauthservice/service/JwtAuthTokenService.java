package org.service.auth.chatappauthservice.service;

import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.io.Decoders;
import io.jsonwebtoken.security.Keys;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.security.Key;
import java.util.Date;

@Service
public class JwtAuthTokenService implements AuthTokenService {

    @Value("${application.jwt.secret}")
    private String secretKey;

    public JwtAuthTokenService() {

    }

    @Override
    public String createToken(UserDTO user, long expiration) {
        Date createdAt = new Date(System.currentTimeMillis());
        Date expireAt = new Date(createdAt.getTime() + expiration);

        return Jwts.builder()
                .header()
                .type("JWT")
                .and()
                .claims()
                .subject(String.valueOf(user.userId()))
                .add("email", user.email())
                .add("username", user.username())
                .add("role", user.role())
                .issuedAt(createdAt)
                .expiration(expireAt)
                .and()
                .signWith(getSignInKey())
                .compact();
    }

    private Key getSignInKey() {
        byte[] keyBytes = Decoders.BASE64.decode(secretKey);
        return Keys.hmacShaKeyFor(keyBytes);
    }

}
