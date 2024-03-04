package org.service.auth.chatappauthservice.service;

import lombok.AllArgsConstructor;
import org.service.auth.chatappauthservice.repository.UserRepository;
import org.springframework.stereotype.Service;

@AllArgsConstructor
@Service
public class UserServiceImpl implements UserService {

    private final UserRepository userRepository;

}
