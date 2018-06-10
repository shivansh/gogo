	.data
x:		.word	0
y:		.word	0
t0:		.word	0
t1:		.word	0
return.0:	.word	0
return.1:	.word	0
newline.0:	.asciiz "\n"
temp.0:		.word	0
temp.1:		.word	0
a.1:		.word	0
b.2:		.word	0

	.text

runtime:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end runtime
temp:
	addi	$sp, $sp, -4
	sw	$ra, 0($sp)
	lw	$3, temp.0	# temp.0 -> $3
	move	$5, $3		# x -> $5
	lw	$3, temp.1	# temp.1 -> $3
	move	$6, $3		# y -> $6
	addi	$3, $5, 1
	addi	$7, $6, 1
	move	$8, $3		# return.0 -> $8
	sw	$8, return.0	# spilled return.0, freed $8
	move	$8, $7		# return.1 -> $8
	# Store dirty variables back into memory
	sw	$3, t0
	sw	$5, x
	sw	$6, y
	sw	$7, t1
	sw	$8, return.1

	lw	$ra, 0($sp)
	addi	$sp, $sp, 4
	jr	$ra
	.end temp

	.globl main
	.ent main
main:
	li	$3, 1		# temp.0 -> $3
	sw	$3, temp.0	# spilled temp.0, freed $3
	li	$3, 2		# temp.1 -> $3
	sw	$3, temp.1
	jal	temp
	lw	$3, temp.1
	lw	$3, return.0	# return.0 -> $3
	move	$5, $3		# a.1 -> $5
	lw	$3, return.1	# return.1 -> $3
	move	$6, $3		# b.2 -> $6
	li	$2, 1
	move	$4, $5
	syscall
	li	$2, 4
	la	$4, newline.0
	syscall
	li	$2, 1
	move	$4, $6
	syscall
	# Store dirty variables back into memory
	sw	$5, a.1
	sw	$6, b.2
	li	$2, 10
	syscall
	.end main
