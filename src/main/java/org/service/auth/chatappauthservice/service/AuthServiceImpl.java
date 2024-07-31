package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import io.jsonwebtoken.Claims;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.configurations.AppConfiguration;
import org.service.auth.chatappauthservice.constants.StatusMessage;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.constants.Role;
import org.service.auth.chatappauthservice.constants.TokenState;
import org.service.auth.chatappauthservice.constants.TokenType;
import org.service.auth.chatappauthservice.exception.authorize.InvalidAuthorizationHeaderException;
import org.service.auth.chatappauthservice.exception.authorize.MissingAccessTokenException;
import org.service.auth.chatappauthservice.exception.refresh.MissingRefreshTokenException;
import org.service.auth.chatappauthservice.exception.refresh.ReusedRefreshTokenException;
import org.service.auth.chatappauthservice.exception.token.ExpiredTokenException;
import org.service.auth.chatappauthservice.exception.token.InvalidTokenException;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.service.auth.chatappauthservice.response.LogoutResponse;
import org.service.auth.chatappauthservice.response.RefreshResponse;
import org.service.auth.chatappauthservice.utils.UserDTOMapper;
import org.springframework.http.*;
import org.springframework.stereotype.Service;
import org.springframework.web.bind.annotation.RequestHeader;

import java.math.BigInteger;
import java.util.Arrays;
import java.util.Map;

@Service
public class AuthServiceImpl implements AuthService {

	private final UserService userService;

	private final AuthTokenService authTokenService;

	private final AppConfiguration configuration;

	public AuthServiceImpl(UserService userService, AuthTokenService authTokenService, AppConfiguration configuration) {
		this.userService = userService;
		this.authTokenService = authTokenService;
		this.configuration = configuration;
	}

	@Override
    public ResponseEntity<AuthenticationResponse> login(JsonNode request)
            throws UserNotFoundException, InvalidUserException {
        configuration.getLogger()
                .debug(STR."[login service]: retrieved request: \{request}");
        JsonNode jsonUser = request.get("payload").get("user");
        String email = jsonUser.get("email").asText();
        String password = jsonUser.get("password").asText();

        configuration.getLogger()
                .debug(STR."[login service]: calling user service to validate user information");
        User user = userService.getValidUser(email, password);
        UserDTOMapper userDTOMapper = new UserDTOMapper();

        configuration.getLogger()
                .debug(STR."[login service]: calling auth token service to create new AT-RT pair");
        String
                accessToken =
                authTokenService.createAccessToken(userDTOMapper.apply(user));
        String
                refreshToken =
                authTokenService.createRefreshToken(userDTOMapper.apply(user));

        configuration.getLogger()
                .debug(STR."[login service]: calling user service to add user's new refresh token");
        userService.addRefreshToken(user.getUserId(), refreshToken);

        configuration.getLogger()
                .debug(STR."[login service]: building response");
        AuthenticationResponse
                responseBody =
                AuthenticationResponse.builder()
                        .accessToken(accessToken)
                        .build();

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.add(HttpHeaders.SET_COOKIE,
                getRefreshTokenCookie(refreshToken).toString());

        return new ResponseEntity<>(responseBody, headers, HttpStatus.OK);
    }

	@Override
    public ResponseEntity<AuthorizationResponse> authorize(
            @RequestHeader Map<String, String> headers)
            throws
            InvalidAuthorizationHeaderException,
            MissingAccessTokenException,
            InvalidTokenException,
            UserNotFoundException {
        configuration.getLogger()
                .debug(STR."[authorization service]: retrieved header: \{headers}");
        if (!headers.containsKey("authorization")) {
            throw new MissingAccessTokenException(StatusMessage.MISSING_CREDENTIAL);
        }

        if (!headers.get("authorization").startsWith("Bearer ")) {
            throw new InvalidAuthorizationHeaderException(StatusMessage.INVALID_FORMAT);
        }

        String token = headers.get("authorization").substring(7).strip();
        configuration.getLogger()
                .debug(STR."[authorization service]: calling auth token service to check token status");
        if (!authTokenService.checkTokenState(token, TokenType.ACCESS_TOKEN)
                .equals(TokenState.VALID)) {
            throw new InvalidTokenException(StatusMessage.INVALID_TOKEN);
        }

        configuration.getLogger()
                .debug(STR."[authorization service]: building response");
        AuthorizationResponse
                response =
                AuthorizationResponse.builder()
                        .message(StatusMessage.AUTHORIZED)
                        .build();
        return new ResponseEntity<>(response, HttpStatus.OK);
    }

	// TODO: hacker keeps sending this "valid" token and the user need to login
	// again each time he/she need to refresh tokens
	// SOLUTION 1: if the token is expired, even if it seems like being hacked, the
	// server just ignores
	@Override
    public ResponseEntity<RefreshResponse> refresh(Map<String, String> cookies) throws
            MissingRefreshTokenException,
            InvalidTokenException,
            JsonProcessingException,
            ReusedRefreshTokenException,
            UserNotFoundException,
            ExpiredTokenException {
        configuration.getLogger()
                .debug(STR."[refresh service]: retrieved cookies: \{cookies}");
        if (!cookies.containsKey("refresh_token")) {
            throw new MissingRefreshTokenException(StatusMessage.MISSING_RT);
        }

        String refreshToken = cookies.get("refresh_token");
        configuration.getLogger()
                .debug("[refresh service]: calling auth token service to check token state");
        TokenState
                state =
                authTokenService.checkTokenState(refreshToken,
                        TokenType.REFRESH_TOKEN);

        // header, payload and signature does not match
        if (state.equals(TokenState.INVALID)) {
            throw new InvalidTokenException(StatusMessage.INVALID_TOKEN);
        }

        configuration.getLogger()
                .debug("[refresh service]: calling auth token service to extract user ID");
        // just in case the token is expired
        Map<String, String> body = authTokenService.parseBody(refreshToken);
        BigInteger userId = BigInteger.valueOf(Long.parseLong(body.get("sub")));

        configuration.getLogger()
                .debug("[refresh service]: calling user service to get user's refresh tokens");
        // get user's all current active refresh tokens
        String[] refreshTokens = userService.getUserRefreshTokens(userId);
        int oldLength = refreshTokens.length;

        refreshTokens = Arrays.stream(refreshTokens)
                .filter(token -> !token.equals(refreshToken))
                .toList()
                .toArray(new String[0]);
        boolean found = refreshTokens.length < oldLength;

        configuration.getLogger()
                .debug("[refresh service]: calling user service to update user's refresh tokens");
        // if not found but expired, follow SOLUTION 1
        // else, remove all active refresh tokens so this user must log in again
        // in all devices to prevent hacker to use stolen refresh token
        userService.updateUserRefreshTokens(userId,
                found || state.equals(TokenState.EXPIRED)
                ? refreshTokens
                : new String[0]);

        // make this user to log in again in this device because the token is
        // expired
        if (state.equals(TokenState.EXPIRED)) {
            throw new ExpiredTokenException(StatusMessage.TOKEN_EXPIRED);
        }

        if (!found) {
            throw new ReusedRefreshTokenException(StatusMessage.INVALID_TOKEN);
        }

        // this token down here must be valid
        configuration.getLogger()
                .debug("[refresh service]: calling auth token service to create new AT-RT pair");
        String
                email =
                authTokenService.extractClaim(refreshToken,
                        claims -> claims.get("email", String.class),
                        TokenType.REFRESH_TOKEN);
        String
                username =
                authTokenService.extractClaim(refreshToken,
                        claims -> claims.get("username", String.class),
                        TokenType.REFRESH_TOKEN);
        Role
                role =
                Role.valueOf(authTokenService.extractClaim(refreshToken,
                        claims -> claims.get("role", String.class),
                        TokenType.REFRESH_TOKEN));
        UserDTO user = new UserDTO(userId, email, username, role);

        String newAccessToken = authTokenService.createAccessToken(user);
        String newRefreshToken = authTokenService.createRefreshToken(user);

        configuration.getLogger()
                .debug("[refresh service]: calling user service to add user's new refresh token");
        userService.addRefreshToken(user.userId(), newRefreshToken);

        configuration.getLogger()
                .debug("[refresh service]: building response");
        RefreshResponse
                response =
                RefreshResponse.builder().accessToken(newAccessToken).build();

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.add(HttpHeaders.SET_COOKIE,
                getRefreshTokenCookie(newRefreshToken).toString());

        return new ResponseEntity<>(response, headers, HttpStatus.OK);
    }

	@Override
    public ResponseEntity<LogoutResponse> logout(Map<String, String> cookies) throws
            InvalidTokenException,
            ExpiredTokenException,
            MissingRefreshTokenException,
            UserNotFoundException {
        configuration.getLogger()
                .debug(STR."[logout service]: retrieved cookies: \{cookies}");
        if (!cookies.containsKey("refresh_token")) {
            throw new MissingRefreshTokenException(StatusMessage.MISSING_RT);
        }

        configuration.getLogger()
                .debug("[logout service]: calling auth token service to check token state");
        String refreshToken = cookies.get("refresh_token");
        TokenState
                state =
                authTokenService.checkTokenState(refreshToken,
                        TokenType.REFRESH_TOKEN);

        switch (state) {
            case INVALID ->
                    throw new InvalidTokenException(StatusMessage.INVALID_TOKEN);
            case EXPIRED ->
                    throw new ExpiredTokenException(StatusMessage.TOKEN_EXPIRED);
            case VALID -> {
            }
            default ->
                    throw new RuntimeException(StatusMessage.UNHANDLED_EXCEPTION);
        }

        // now the refresh token must be valid, add the current refresh token to the black
        // list
        configuration.getLogger()
                .debug("[logout service]: calling auth token service to extract user ID from RT");
        BigInteger
                userId =
                BigInteger.valueOf(Long.parseLong(authTokenService.extractClaim(
                        refreshToken,
                        Claims::getSubject,
                        TokenType.REFRESH_TOKEN)));
        configuration.getLogger()
                .debug(STR."[logout service]: calling user token service to remove used RT: \{refreshToken}");

        String[]
                refreshTokens =
                Arrays.stream(userService.getUserRefreshTokens(userId))
                        .filter(token -> !token.equals(refreshToken))
                        .toList()
                        .toArray(new String[0]);
        userService.updateUserRefreshTokens(userId, refreshTokens);

        configuration.getLogger()
                .debug(STR."[logout service]: building response");
        LogoutResponse
                response =
                LogoutResponse.builder().message("OK").build();

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.add(HttpHeaders.SET_COOKIE,
                getRefreshTokenCookie(null).toString()); // clear cookie

        return new ResponseEntity<>(response, headers, HttpStatus.OK);
    }

	// if you want to clear cookie then set refreshToken to null
	private ResponseCookie getRefreshTokenCookie(String refreshToken) {
        if (refreshToken == null) {
            return ResponseCookie.from("refresh_token")
                    .maxAge(0)
                    .httpOnly(true)
                    .path(STR."/api/\{configuration.getApiVersion()}/auth")
                    .build();
        }

        return ResponseCookie.from("refresh_token", refreshToken)
                // .secure(true) //TODO: add this when you have https
                .maxAge(configuration.getRTLifespanInMs() / 1000)
                .httpOnly(true)
                .path(STR."/api/\{configuration.getApiVersion()}/auth")
                .build();
    }

}
