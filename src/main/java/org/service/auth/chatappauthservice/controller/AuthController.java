package org.service.auth.chatappauthservice.controller;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import lombok.AllArgsConstructor;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.service.auth.chatappauthservice.response.RefreshResponse;
import org.service.auth.chatappauthservice.service.AuthService;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api/v1/auth")
@AllArgsConstructor
public class AuthController implements AuthApi {

	private static final Logger logger = LogManager.getLogger(AuthController.class);

	private final AuthService authService;

	@Override
    public ResponseEntity<AuthenticationResponse> authenticate(@RequestBody JsonNode request) {
        logger.info(STR."Received request: \{request}");
        return authService.login(request);
    }

	@Override
    public ResponseEntity<AuthorizationResponse> authorize(@RequestHeader Map<String, String> headers) {
        logger.info(STR."Received headers: \{headers}");
        return authService.authorize(headers);
    }

	@Override
    public ResponseEntity<RefreshResponse> refresh(HttpServletRequest request) throws JsonProcessingException {
        Cookie[] rawCookies = request.getCookies();
        Map<String, String> cookies = new HashMap<>();

        if (rawCookies != null) {
            Arrays.stream(rawCookies).forEach(rawCookie ->
                    cookies.merge(rawCookie.getName(), rawCookie.getValue(), (before, after) -> after));
        }

        logger.info(STR."Received cookies: \{cookies}");
        return authService.refresh(cookies);
    }

}
