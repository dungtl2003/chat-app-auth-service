package org.service.auth.chatappauthservice.DTO;

import org.service.auth.chatappauthservice.constants.Role;

public record UserDTO(String userId, String email, String username, Role role) {
}
