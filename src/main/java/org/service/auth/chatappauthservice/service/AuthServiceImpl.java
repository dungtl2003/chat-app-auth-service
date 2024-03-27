package org.service.auth.chatappauthservice.service;

import com.fasterxml.jackson.databind.JsonNode;
import lombok.AllArgsConstructor;
import org.service.auth.chatappauthservice.entity.User;
import org.service.auth.chatappauthservice.exception.client.InvalidUserException;
import org.service.auth.chatappauthservice.exception.client.UserNotFoundException;
import org.springframework.http.*;
import org.springframework.stereotype.Service;

@AllArgsConstructor
@Service
public class AuthServiceImpl implements AuthService {

	private final UserService userService;

	private final AuthTokenService authTokenService;

	@Override
    public ResponseEntity<String> authenticate(JsonNode request) throws UserNotFoundException, InvalidUserException {
        JsonNode jsonUser = request.get("payload").get("user");
        String email = jsonUser.get("email").asText();
        String password = jsonUser.get("password").asText();

        User user = userService.getValidUser(email, password);

        String accessToken = authTokenService.createAccessToken(user);
        String refreshToken = authTokenService.createRefreshToken(user);

        userService.updateUserRefreshTokens(user, refreshToken);

        HttpHeaders headers = new HttpHeaders();
        headers.setContentType(MediaType.APPLICATION_JSON);
        headers.add(HttpHeaders.AUTHORIZATION, STR."Bearer \{accessToken}");

        ResponseCookie cookie = ResponseCookie.from("jwt", refreshToken)
//                .secure(true) //TODO: add this when you have https
                .maxAge(7 * 24 * 60 * 60) //same as the refresh token
                .httpOnly(true)
                .build();

        headers.add(HttpHeaders.SET_COOKIE, cookie.toString());

        return new ResponseEntity<>(headers, HttpStatus.OK);
    }

}
