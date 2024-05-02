package org.service.auth.chatappauthservice.controller;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.enums.ParameterIn;
import io.swagger.v3.oas.annotations.media.Content;
import io.swagger.v3.oas.annotations.media.ExampleObject;
import io.swagger.v3.oas.annotations.media.Schema;
import io.swagger.v3.oas.annotations.responses.ApiResponse;
import io.swagger.v3.oas.annotations.security.SecurityRequirement;
import io.swagger.v3.oas.annotations.security.SecurityScheme;
import io.swagger.v3.oas.annotations.security.SecuritySchemes;
import jakarta.servlet.http.Cookie;
import jakarta.servlet.http.HttpServletRequest;
import lombok.AllArgsConstructor;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.service.auth.chatappauthservice.exception.authorize.InvalidAuthorizationHeaderException;
import org.service.auth.chatappauthservice.exception.authorize.MissingAccessTokenException;
import org.service.auth.chatappauthservice.exception.refresh.MissingRefreshTokenException;
import org.service.auth.chatappauthservice.exception.token.InvalidTokenException;
import org.service.auth.chatappauthservice.exception.user.InvalidUserException;
import org.service.auth.chatappauthservice.exception.user.UserNotFoundException;
import org.service.auth.chatappauthservice.response.AuthenticationResponse;
import org.service.auth.chatappauthservice.response.AuthorizationResponse;
import org.service.auth.chatappauthservice.response.RefreshResponse;
import org.service.auth.chatappauthservice.service.AuthService;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Component;
import org.springframework.web.bind.annotation.*;

import java.util.Arrays;
import java.util.HashMap;
import java.util.Map;

@RestController
@AllArgsConstructor
@RequestMapping("/api/v1/auth")
public class AuthController {

	private static final Logger logger = LogManager.getLogger(AuthController.class);

	private final AuthService authService;

	@Operation(
            summary = "Logs user into the system",
            responses = {
                    @ApiResponse(responseCode="200", description="Success"),
            },
            requestBody = @io.swagger.v3.oas.annotations.parameters.RequestBody(required = true,
                    content = @Content(
                            examples = @ExampleObject(
                                    value = """
                                            {
                                                "payload" : {
                                                    "user" : {
                                                        "email": "toan@gmail.com",
                                                        "password" : "12345678"
                                                    }
                                                }
                                            }
                                            """)
            ))
    )
	@PostMapping(value = "/login", consumes = "application/json")
    public ResponseEntity<AuthenticationResponse> authenticate(@RequestBody JsonNode request) throws UserNotFoundException, InvalidUserException {
        logger.info(STR."Received request: \{request}");
        return authService.login(request);
    }

	@Operation(
            summary = "Check user's access token validation",
            responses = {
                    @ApiResponse(responseCode="200", description="Authorized"),
            },
            security = @SecurityRequirement(name="authorization")
    )
	@GetMapping(value="/authorize")
    public ResponseEntity<AuthorizationResponse> authorize(@RequestHeader Map<String, String> headers)
            throws MissingAccessTokenException, InvalidAuthorizationHeaderException,
            InvalidTokenException {
        logger.info(STR."Received headers: \{headers}");
        return authService.authorize(headers);
    }

	@Operation(
            summary = "Refresh user's token",
            responses = {
                    @ApiResponse(responseCode="200", description="Success"),
            },
            security = @SecurityRequirement(name="refresh_token")
    )
	@GetMapping("/refresh")
    public ResponseEntity<RefreshResponse> refresh(HttpServletRequest request)
            throws MissingRefreshTokenException, InvalidTokenException,
            JsonProcessingException {
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
