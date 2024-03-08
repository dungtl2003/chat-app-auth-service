package org.service.auth.chatappauthservice.DTO;

public record UserDTO(Long userId, String email, String username, String password, String role) {
}
