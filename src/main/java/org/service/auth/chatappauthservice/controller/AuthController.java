package org.service.auth.chatappauthservice.controller;

import com.fasterxml.jackson.databind.JsonNode;
import lombok.AllArgsConstructor;
import org.apache.logging.log4j.LogManager;
import org.apache.logging.log4j.Logger;
import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.exception.client.InvalidUserException;
import org.service.auth.chatappauthservice.exception.client.UserNotFoundException;
import org.service.auth.chatappauthservice.service.AuthTokenService;
import org.service.auth.chatappauthservice.service.UserService;
import org.springframework.http.HttpHeaders;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@AllArgsConstructor
@RequestMapping("/api/v1/auth")
public class AuthController {

    private static final Logger logger = LogManager.getLogger(AuthController.class);

    private final UserService userService;

    private final AuthTokenService authTokenService;

    @GetMapping("/bcrypt")
    public void createRandomUser(@RequestBody JsonNode request) {
        JsonNode payload = request.get("payload");
        logger.debug(STR."payload: \{request}");
    }

    @PostMapping("/authenticate")
    public ResponseEntity<String> authenticate(@RequestBody JsonNode request) throws UserNotFoundException, InvalidUserException {
        logger.info(STR."Received request: \{request}");
        JsonNode jsonUser = request.get("payload").get("user");
        String email = jsonUser.get("email").asText();
        String password = jsonUser.get("password").asText();

        UserDTO user = userService.getValidUser(email, password);
        long expiration = 24 * 3600; //1 day
        String token = authTokenService.createToken(user, expiration);

        HttpHeaders headers = new HttpHeaders();
        headers.add("Authorization", STR."Bearer \{token}");

        return new ResponseEntity<>(headers, HttpStatus.OK);
    }

}
