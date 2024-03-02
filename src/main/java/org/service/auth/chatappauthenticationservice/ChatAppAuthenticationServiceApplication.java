package org.service.auth.chatappauthenticationservice;

import org.service.auth.chatappauthenticationservice.configurations.PostgresqlConfigProperties;
import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.EnableConfigurationProperties;

@SpringBootApplication
@EnableConfigurationProperties(PostgresqlConfigProperties.class)
public class ChatAppAuthenticationServiceApplication {

    public static void main(String[] args) {
        SpringApplication.run(ChatAppAuthenticationServiceApplication.class, args);
    }
}
