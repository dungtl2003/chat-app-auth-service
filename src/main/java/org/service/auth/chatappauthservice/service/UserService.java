package org.service.auth.chatappauthservice.service;

import org.service.auth.chatappauthservice.DTO.UserDTO;
import org.service.auth.chatappauthservice.entity.User;

public interface UserService {

    UserDTO getValidUser(String email, String password);

    void add(User user);

    void delete(Long userId);
}
