import {
  Body,
  Button,
  Container,
  Head,
  Html,
  Preview,
  Section,
  Tailwind,
  Text,
} from "@react-email/components";
import * as React from "react";
import { tailwindConfig } from "../tailwind.config";

// Go template placeholders
const NAME = "{{.Name}}";
const LOGIN_LINK = "{{.LoginLink}}";

export const Welcome = () => {
  return (
    <Html>
      <Head />
      <Preview>Bem-vindo ao ZapAgenda!</Preview>
      <Tailwind config={tailwindConfig}>
        <Body className="bg-gray-100 font-sans">
          <Container className="bg-white mx-auto my-10 max-w-xl rounded-lg shadow-sm">
            <Section className="px-12 py-8 border-b border-gray-200">
              <Text className="text-2xl font-bold text-brand m-0">
                ZapAgenda
              </Text>
            </Section>

            <Section className="px-12 py-8">
              <Text className="text-xl font-bold text-brand mb-6">
                Bem-vindo ao ZapAgenda!
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-4">
                Olá {NAME},
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-4">
                Sua conta foi criada com sucesso! Estamos muito felizes em ter
                você conosco.
              </Text>

              <Text className="text-base text-gray-600 leading-7 mb-6">
                O ZapAgenda vai ajudar você a gerenciar seus agendamentos de
                forma simples e eficiente. Comece agora mesmo:
              </Text>

              <Button
                href={LOGIN_LINK}
                className="bg-brand text-white font-semibold py-3 px-6 rounded-lg"
              >
                Acessar Minha Conta
              </Button>

              <Text className="text-base text-gray-600 leading-7 mt-6 mb-4">
                Se tiver alguma dúvida, não hesite em entrar em contato com
                nosso suporte.
              </Text>

              <Text className="text-base text-gray-600 leading-7">
                Atenciosamente,
                <br />
                Equipe ZapAgenda
              </Text>
            </Section>

            <Section className="px-12 py-6 border-t border-gray-200">
              <Text className="text-xs text-gray-400 text-center m-0">
                © {new Date().getFullYear()} ZapAgenda. Todos os direitos
                reservados.
              </Text>
            </Section>
          </Container>
        </Body>
      </Tailwind>
    </Html>
  );
};

export default Welcome;
