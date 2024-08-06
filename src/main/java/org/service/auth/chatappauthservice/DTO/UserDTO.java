package org.service.auth.chatappauthservice.DTO;

import org.service.auth.chatappauthservice.constants.Role;

import java.math.BigInteger;

public record UserDTO(BigInteger userId, String email, String username, Role role) {
}
