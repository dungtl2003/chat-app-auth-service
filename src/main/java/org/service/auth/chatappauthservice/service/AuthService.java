package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
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
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RequestHeader;

import java.util.Map;

public interface AuthService {

	// @formatter:off
    /**
     * check given email and password
     *
     * @param request the request body, structure:
     * <pre>
     * {@code
     * {
     * 	user: {
     * 		email: abc,
     * 		password: def
     * 	}
     * }
     * }
     * </pre>
     * @return `200` if the request success<br> `404` if it gets `UserNotFoundException`<br> `401` if it gets `InvalidUserException`
     * @throws UserNotFoundException if it cannot find given email in database
     * @throws InvalidUserException  if the given email and password do not match to email - password in database
     */
    ResponseEntity<AuthenticationResponse> login(@RequestBody JsonNode request)
            throws UserNotFoundException, InvalidUserException;

    /**
     * authorize the request with given access token
     *
     * @param headers header of the request, must contain `Authorization: Bearer token`
     * @return `200` if the request success<br> `401` if it gets `MissingAccessTokenException`, `InvalidAuthorizationHeaderException`, `InvalidTokenException`<br>`404` if it gets `UserNotFoundException`
     * @throws MissingAccessTokenException if the header is missing `Authorization: Bearer token`
     * @throws InvalidAuthorizationHeaderException if the format of the header is incorrect
     * @throws InvalidTokenException if the given access token is invalid
     * @throws UserNotFoundException if the user data in token does not exist in database. THIS SHOULD NOT BE HAPPENED
     */
    ResponseEntity<AuthorizationResponse> authorize(
            @RequestHeader Map<String, String> headers)
            throws
            MissingAccessTokenException,
            InvalidAuthorizationHeaderException,
            InvalidTokenException,
            UserNotFoundException;

    /**
     * try to refresh new access - refresh token pair using given refresh token
     * TODO: right now, each RT should belong to a device, but all RTs are inside a list. Need a table to manage devices, each one has a RT
     *
     * @param cookies must have `refresh_token` cookie
     * @return `200` if the request success<br> `401` if it gets `ExpiredTokenException`, `MissingRefreshTokenException`, `InvalidTokenException`, `ReusedRefreshTokenException`<br>`404` if it gets `UserNotFoundException`
     * @throws MissingRefreshTokenException if the cookie does not contain refresh token
     * @throws InvalidTokenException if the refresh token is invalid
     * @throws JsonProcessingException if the refresh token is not in json format. THIS SHOULD NOT BE HAPPENED
     * @throws ReusedRefreshTokenException if this refresh token has been used before
     * @throws UserNotFoundException if the user data in token does not exist in database. THIS SHOULD NOT BE HAPPENED
     * @throws ExpiredTokenException if the token is expired
     */
    ResponseEntity<RefreshResponse> refresh(
            @RequestHeader Map<String, String> cookies)
            throws
            MissingRefreshTokenException,
            InvalidTokenException,
            JsonProcessingException,
            ReusedRefreshTokenException,
            UserNotFoundException,
            ExpiredTokenException;

    /**
     * logout user by adding refresh token to the black list, plus clearing refresh token cookie
     *
     * @param cookies must have `refresh_token` cookie
     * @return `200` if the request success<br> `401` if it gets `ExpiredTokenException`, `MissingRefreshTokenException`, `InvalidTokenException`, <br>`404` if it gets `UserNotFoundException`
     * @throws InvalidTokenException if the refresh token is invalid
     * @throws ExpiredTokenException if the token is expired
     * @throws MissingRefreshTokenException if the cookie does not contain refresh token
     * @throws UserNotFoundException if the user data in token does not exist in database. THIS SHOULD NOT BE HAPPENED
     */
    ResponseEntity<LogoutResponse> logout(@RequestHeader Map<String, String> cookies) throws
    InvalidTokenException,
    ExpiredTokenException,
    MissingRefreshTokenException,
    UserNotFoundException;
}
