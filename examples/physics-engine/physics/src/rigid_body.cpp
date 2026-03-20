#include <physics/rigid_body.h>

namespace physics {

RigidBody::RigidBody(float mass) : mass(mass) {}

void RigidBody::apply_force(const mathcore::Vec3& force) {
    acceleration = acceleration + force * (1.0f / mass);
}

void RigidBody::step(float dt) {
    velocity = velocity + acceleration * dt;
    position = position + velocity * dt;
    acceleration = mathcore::Vec3();
}

}
