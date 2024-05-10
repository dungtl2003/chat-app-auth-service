package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import io.jsonwebtoken.Claims;
import io.jsonwebtoken.JwtException;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.entity.enums.TokenState;
import org.service.auth.chatappauthservice.entity.enums.TokenType;

import java.util.Map;
import java.util.function.Function;

public interface AuthTokenService {

	/**
	 * Create and return access token with default recommended access token's expiration
	 * time.
	 * @param user The user you want to authenticate.
	 * @return The user's access token.
	 */
	String createAccessToken(UserDTO user);

	/**
	 * Create and return refresh token with default recommended refresh token's expiration
	 * time.
	 * @param user The user you want to authenticate.
	 * @return The user's refresh token.
	 */
	String createRefreshToken(UserDTO user);

	/**
	 * Create and return access token, same as {@link #createAccessToken(UserDTO)} but
	 * with custom expiration time.
	 * @param expiration The token's expiration date in milliseconds.
	 * @return The user's access token.
	 */
	String createAccessToken(UserDTO user, long expiration);

	/**
	 * Create and return refresh token, same as {@link #createRefreshToken(UserDTO)} but
	 * with custom expiration time.
	 * @param expiration The token's expiration date in milliseconds.
	 * @return The user's refresh token.
	 */
	String createRefreshToken(UserDTO user, long expiration);

	/**
	 * Check the state of the given token.
	 * @param token The {@code token} you want to check.
	 * @param type The {@code type} of the token.
	 * @see TokenType
	 */
	TokenState checkTokenState(String token, TokenType type);

	/**
	 * Extract specific token's claim. This is used for JWS, you also need to point out
	 * the {@code type} of the token, so the function can decode it.
	 * @param token The {@code token} you want to decode.
	 * @param claimsResolver The {@code function} to resolve the needed claim.
	 * @param type The {@code type} of the token.
	 * @param <T> The {@code type} of the result function.
	 * @return The specific claim.
	 * @throws JwtException If the token is invalid. This can be either the header and
	 * payload does not match the signature or the token is expired.
	 * @see TokenType
	 */
	<T> T extractClaim(String token, Function<Claims, T> claimsResolver, TokenType type) throws JwtException;

	/**
	 * Extract token's body. This function will decode the body without security feature.
	 * It is generally <b>recommended</b> to use
	 * {@link #extractClaim(String, Function, TokenType)} instead for security feature,
	 * and you can get the exact claim you want.
	 * @param token The {@code token} you want to extract the body from.
	 * @return All key-value pairs of the body.
	 * @throws JsonProcessingException When token's body is not in json form.
	 */
	Map<String, String> parseBody(String token) throws JsonProcessingException;

}
